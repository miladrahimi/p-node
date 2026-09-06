package xray

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/util"
	xc "github.com/miladrahimi/p-node/pkg/xray/config"
	xrayUtil "github.com/miladrahimi/p-node/pkg/xray/util"
	stats "github.com/xtls/xray-core/app/stats/command"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ErrPortConflict is returned by Reconfigure when the inbound ports of the new config are not usable.
var ErrPortConflict = errors.New("xray: port conflict")

// apiTag is the tag of the inbound that serves the Xray gRPC API.
const apiTag = "api"

// Xray represents the Xray instance which is running in the background.
type Xray struct {
	l          *logger.Logger
	config     *xc.Config
	configPath string
	binaryPath string
	command    *exec.Cmd
	exited     chan struct{} // closed once the current process has exited
	killed     *atomic.Bool  // set before the current process is killed on purpose
	connection *grpc.ClientConn
	locker     sync.Mutex
	context    context.Context
}

// New creates a new Xray instance.
func New(c context.Context, logger *logger.Logger, logLevel, configPath, binaryPath string) *Xray {
	return &Xray{
		context:    c,
		l:          logger,
		config:     xc.New(logLevel),
		binaryPath: binaryPath,
		configPath: configPath,
	}
}

// Load loads the stored configuration if it already exists and picks the API port for this process.
func (x *Xray) Load() error {
	x.locker.Lock()
	defer x.locker.Unlock()

	if util.FileExist(x.configPath) {
		if err := x.loadConfig(); err != nil {
			return errors.WithStack(err)
		}
	}

	// The API inbound listens on a random free port chosen once per process;
	// Reconfigure keeps it, so pushed and pulled configs never carry it.
	if api := x.config.FindInbound(apiTag); api != nil {
		port, err := util.FreePort()
		if err != nil {
			return errors.WithStack(err)
		}
		api.Port = port
	}

	return errors.WithStack(x.saveConfig())
}

// Stop kills the Xray instance.
func (x *Xray) Stop() error {
	x.locker.Lock()
	defer x.locker.Unlock()

	return x.stopLocked()
}

// Run runs the Xray proxy instance in the background.
func (x *Xray) Run() error {
	x.locker.Lock()
	defer x.locker.Unlock()

	return x.runLocked()
}

// Connect creates the Xray API client if it does not exist yet.
// The gRPC client dials lazily, so this never blocks on the Xray process.
func (x *Xray) Connect() error {
	x.locker.Lock()
	defer x.locker.Unlock()

	if x.connection != nil {
		return nil
	}

	inbound := x.config.FindInbound(apiTag)
	if inbound == nil {
		return errors.New("xray: no api inbound")
	}

	address := "127.0.0.1:" + strconv.Itoa(inbound.Port)
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.WithStack(err)
	}

	x.connection = conn
	x.l.Debug("xray: api client created", zap.String("address", address))
	return nil
}

// Reconfigure applies the given config and restarts Xray if it differs from the running one.
// The API inbound port of the running instance is kept, so callers need not know it.
func (x *Xray) Reconfigure(newConfig *xc.Config) error {
	if newConfig == nil {
		return errors.New("xray: config is nil")
	}
	if err := newConfig.Validate(); err != nil {
		return errors.WithStack(err)
	}

	x.locker.Lock()
	defer x.locker.Unlock()

	if newApi := newConfig.FindInbound(apiTag); newApi != nil {
		if currentApi := x.config.FindInbound(apiTag); currentApi != nil {
			newApi.Port = currentApi.Port
		} else {
			port, err := util.FreePort()
			if err != nil {
				return errors.WithStack(err)
			}
			newApi.Port = port
		}
	}

	if x.config.Equals(newConfig) {
		x.l.Debug("xray: config unchanged, restart skipped")
		return nil
	}

	if err := validatePorts(x.config, newConfig, util.PortFree); err != nil {
		return err
	}

	x.config = newConfig
	return x.restartLocked()
}

// QueryStats queries the Xray stats and resets the counters.
func (x *Xray) QueryStats() ([]*stats.Stat, error) {
	if err := x.Connect(); err != nil {
		return nil, errors.WithStack(err)
	}

	x.locker.Lock()
	conn := x.connection
	x.locker.Unlock()
	if conn == nil {
		return nil, errors.New("xray: api connection is not established")
	}

	ctx, cancel := context.WithTimeout(x.context, 10*time.Second)
	defer cancel()

	client := stats.NewStatsServiceClient(conn)
	qs, err := client.QueryStats(ctx, &stats.QueryStatsRequest{Reset_: true}, grpc.WaitForReady(true))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return qs.GetStat(), nil
}

// GenerateX25519 generates an X25519 key pair (private and public) using the Xray binary.
func (x *Xray) GenerateX25519() (string, string, error) {
	if !util.FileExist(x.binaryPath) {
		return "", "", errors.New("xray: binary not found")
	}

	cmd := exec.Command(x.binaryPath, "x25519")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.WithMessage(err, "xray: x25519 failed: "+strings.TrimSpace(string(output)))
	}

	privateKey, password, err := xrayUtil.ParseX25519Output(output)

	return privateKey, password, errors.WithStack(err)
}

// validatePorts checks that the inbound ports of next are unique and either free or
// currently held by the running instance described by current. The API inbound is
// managed by Xray itself and is only checked for uniqueness.
func validatePorts(current, next *xc.Config, portFree func(int) bool) error {
	held := map[int]int{}
	if current != nil {
		for _, in := range current.Inbounds {
			if in.Tag != apiTag && in.Port != 0 {
				held[in.Port]++
			}
		}
	}

	seen := map[int]bool{}
	apiCount := 0
	for _, in := range next.Inbounds {
		if in.Tag == apiTag {
			if apiCount++; apiCount > 1 {
				return errors.Wrap(ErrPortConflict, "only one api inbound is allowed")
			}
			continue
		}
		if in.Port == 0 {
			continue
		}
		if seen[in.Port] {
			return errors.Wrapf(ErrPortConflict, "port %d is used by multiple inbounds", in.Port)
		}
		seen[in.Port] = true

		if held[in.Port] > 0 {
			held[in.Port]--
			continue
		}
		if !portFree(in.Port) {
			return errors.Wrapf(ErrPortConflict, "port %d for '%s' is already in use", in.Port, in.Tag)
		}
	}

	return nil
}

func (x *Xray) stopLocked() error {
	x.l.Debug("xray: stopping...")

	if x.connection != nil {
		x.l.Debug("xray: closing the api connection...")
		if err := x.connection.Close(); err != nil {
			x.l.Debug("xray: cannot close the api connection", zap.Error(errors.WithStack(err)))
		} else {
			x.l.Debug("xray: the api connection closed")
		}
		x.connection = nil
	}

	if x.command != nil && x.command.Process != nil {
		x.l.Debug("xray: killing the process...")
		x.killed.Store(true)
		if err := x.command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return errors.WithStack(err)
		}
		// Wait for the process to release its ports before a successor is started.
		select {
		case <-x.exited:
			x.l.Debug("xray: the process exited")
		case <-time.After(10 * time.Second):
			x.l.Error("xray: the process did not exit after kill")
		}
	}
	x.command = nil

	x.l.Info("xray: closed")
	return nil
}

func (x *Xray) runLocked() error {
	if !util.FileExist(x.binaryPath) {
		x.l.Fatal("xray: binary not found", zap.String("path", x.binaryPath))
		return errors.New("xray: binary not found")
	}

	if err := x.saveConfig(); err != nil {
		return errors.WithStack(err)
	}

	x.l.Debug("xray: running...")
	cmd := exec.Command(x.binaryPath, "-c", x.configPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	x.l.Info("xray: executing the binary...", zap.String("path", x.binaryPath))
	if err := cmd.Start(); err != nil {
		return errors.WithStack(err)
	}

	exited := make(chan struct{})
	killed := &atomic.Bool{}
	x.command, x.exited, x.killed = cmd, exited, killed

	go func() {
		err := cmd.Wait()
		close(exited)
		if !killed.Load() {
			x.l.Fatal("xray: process exited unexpectedly", zap.Error(errors.WithStack(err)))
		}
	}()

	return nil
}

func (x *Xray) restartLocked() error {
	x.l.Info("xray: restarting...")

	if err := x.stopLocked(); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(x.runLocked())
}

// loadConfig loads the Xray config file.
func (x *Xray) loadConfig() error {
	x.l.Debug("xray: loading config file...")

	content, err := os.ReadFile(x.configPath)
	if err != nil {
		return errors.WithStack(err)
	}

	var newConfig xc.Config
	if err = json.Unmarshal(content, &newConfig); err != nil {
		return errors.WithStack(err)
	}

	if err = newConfig.Validate(); err != nil {
		return errors.WithStack(err)
	}

	x.config = &newConfig
	x.l.Debug("xray: config file loaded")
	return nil
}

// saveConfig saves the Xray config file.
func (x *Xray) saveConfig() error {
	x.l.Debug("xray: saving config file...")

	content, err := json.Marshal(x.config)
	if err != nil {
		return errors.WithStack(err)
	}

	if err = os.MkdirAll(filepath.Dir(x.configPath), 0o755); err != nil {
		return errors.WithStack(err)
	}

	err = os.WriteFile(x.configPath, content, 0644)
	if err == nil {
		x.l.Debug("xray: config file saved")
	}
	return errors.WithStack(err)
}
