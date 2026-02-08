package xray

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/util"
	xc "github.com/miladrahimi/p-node/pkg/xray/config"
	stats "github.com/xtls/xray-core/app/stats/command"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Xray represents the Xray instance which is running in the background.
type Xray struct {
	l          *logger.Logger
	config     *xc.Config
	configPath string
	binaryPath string
	command    *exec.Cmd
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

// Init initializes the Xray instance.
func (x *Xray) Init() error {
	x.locker.Lock()
	defer x.locker.Unlock()

	if err := x.saveConfig(); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(x.loadConfig())
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

// Restart restarts the Xray instance.
func (x *Xray) Restart() error {
	x.locker.Lock()
	defer x.locker.Unlock()

	return x.restartLocked()
}

// Connect connects to the Xray API.
func (x *Xray) Connect() error {
	x.locker.Lock()
	if x.connection != nil {
		x.locker.Unlock()
		return nil
	}
	currentConfig := x.config
	x.locker.Unlock()

	x.l.Debug("xray: connecting to api...")

	inbound := currentConfig.FindInbound("api")
	if inbound == nil {
		return errors.New("no api inbound")
	}

	ctx, cancel := context.WithTimeout(x.context, 10*time.Second)
	defer cancel()

	address := "127.0.0.1:" + strconv.Itoa(inbound.Port)

	for {
		select {
		case <-ctx.Done():
			return errors.New("connection to xray api timed out")
		default:
			time.Sleep(time.Second)
			conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				x.l.Debug("xray: trying to Connect to api", zap.Error(errors.WithStack(err)))
			} else {
				x.locker.Lock()
				x.connection = conn
				x.locker.Unlock()
				x.l.Debug("xray: connected to api successfully")
				return nil
			}
		}
	}
}

// Config returns the Xray config.
func (x *Xray) Config() *xc.Config {
	x.locker.Lock()
	defer x.locker.Unlock()

	return x.config
}

// Reconfigure sets the Xray config.
func (x *Xray) Reconfigure(newConfig *xc.Config) error {
	if newConfig == nil {
		return errors.New("xray: config is nil")
	}
	if err := newConfig.Validate(); err != nil {
		return errors.WithStack(err)
	}

	x.locker.Lock()
	defer x.locker.Unlock()

	x.config = newConfig
	return x.restartLocked()
}

// QueryStats queries the Xray stats.
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

	client := stats.NewStatsServiceClient(conn)
	qs, err := client.QueryStats(context.Background(), &stats.QueryStatsRequest{Reset_: true})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return qs.GetStat(), nil
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
		if err := x.command.Process.Kill(); err != nil {
			return errors.WithStack(err)
		} else {
			x.l.Debug("xray: the process killed")
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
	x.command = exec.Command(x.binaryPath, "-c", x.configPath)
	x.command.Stderr = os.Stderr
	x.command.Stdout = os.Stdout

	x.l.Info("xray: executing the binary...", zap.String("path", x.binaryPath))
	if err := x.command.Start(); err != nil {
		x.command = nil
		return errors.WithStack(err)
	}

	go func(cmd *exec.Cmd) {
		if err := cmd.Wait(); err != nil && err.Error() != "signal: killed" {
			x.l.Fatal("xray: process exited unexpectedly", zap.Error(errors.WithStack(err)))
		}
	}(x.command)

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

	if !util.FileExist(x.configPath) {
		x.l.Debug("xray: no config file found, it is fresh")
		return nil
	}

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
