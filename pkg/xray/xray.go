package xray

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/utils"
	"github.com/miladrahimi/p-node/pkg/logger"
	stats "github.com/xtls/xray-core/app/stats/command"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Xray represents the Xray instance which is running in the background.
type Xray struct {
	l          *logger.Logger
	config     *Config
	configPath string
	binaryPath string
	command    *exec.Cmd
	connection *grpc.ClientConn
	locker     *sync.Mutex
	context    context.Context
}

// New creates a new Xray instance.
func New(c context.Context, logger *logger.Logger, logLevel, configPath, binaryPath string) *Xray {
	return &Xray{
		context:    c,
		l:          logger,
		config:     NewConfig(logLevel),
		binaryPath: binaryPath,
		configPath: configPath,
		locker:     &sync.Mutex{},
	}
}

// Init initializes the Xray instance.
func (x *Xray) Init() error {
	if err := x.saveConfig(); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(x.loadConfig())
}

// Stop kills the Xray instance.
func (x *Xray) Stop() error {
	x.l.Debug("xray: stopping...")

	x.locker.Lock()
	defer x.locker.Unlock()

	if x.connection != nil {
		x.l.Debug("xray: closing the api connection...")
		if err := x.connection.Close(); err != nil {
			x.l.Debug("xray: cannot close the api connection", zap.Error(errors.WithStack(err)))
		} else {
			x.l.Debug("xray: the api connection closed")
		}
	}

	if x.command != nil && x.command.Process != nil {
		x.l.Debug("xray: killing the process...")
		if err := x.command.Process.Kill(); err != nil {
			return errors.WithStack(err)
		} else {
			x.l.Debug("xray: the process killed")
		}
	}

	x.l.Info("xray: closed")
	return nil
}

// Run runs the Xray proxy instance in the background.
func (x *Xray) Run() error {
	if !utils.FileExist(x.binaryPath) {
		x.l.Fatal("xray: binary not found", zap.String("path", x.binaryPath))
		return errors.New("xray: binary not found")
	}

	x.l.Debug("xray: running...")
	x.command = exec.Command(x.binaryPath, "-c", x.configPath)
	x.command.Stderr = os.Stderr
	x.command.Stdout = os.Stdout

	x.l.Info("xray: executing the binary...", zap.String("path", x.binaryPath))
	if err := x.command.Start(); err != nil {
		return errors.WithStack(err)
	}

	go func() {
		if err := x.command.Wait(); err != nil && err.Error() != "signal: killed" {
			x.l.Fatal("xray: process exited unexpectedly", zap.Error(errors.WithStack(err)))
		}
	}()

	return nil
}

// Restart restarts the Xray instance.
func (x *Xray) Restart() {
	x.l.Info("xray: restarting...")

	if err := x.Stop(); err != nil {
		x.l.Error("xray: cannot close", zap.Error(errors.WithStack(err)))
	}

	if err := x.Run(); err != nil {
		x.l.Fatal("xray: cannot run again", zap.Error(errors.WithStack(err)))
	}
}

// Connect connects to the Xray API.
func (x *Xray) Connect() error {
	x.l.Debug("xray: connecting to api...")

	inbound := x.config.FindInbound("api")
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
			var err error
			x.connection, err = grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				x.l.Debug("xray: trying to Connect to api", zap.Error(errors.WithStack(err)))
			} else {
				x.l.Debug("xray: connected to api successfully")
				return nil
			}
		}
	}
}

// Config returns the Xray config.
func (x *Xray) Config() *Config {
	return x.config
}

// SetConfig sets the Xray config.
func (x *Xray) SetConfig(config *Config) {
	x.config = config
}

// QueryStats queries the Xray stats.
func (x *Xray) QueryStats() ([]*stats.Stat, error) {
	client := stats.NewStatsServiceClient(x.connection)
	qs, err := client.QueryStats(context.Background(), &stats.QueryStatsRequest{Reset_: true})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return qs.GetStat(), nil
}

// loadConfig loads the Xray config file.
func (x *Xray) loadConfig() error {
	x.l.Debug("xray: loading config file...")

	if !utils.FileExist(x.configPath) {
		x.l.Debug("xray: no config file found, it is fresh")
		return nil
	}

	content, err := os.ReadFile(x.configPath)
	if err != nil {
		return errors.WithStack(err)
	}

	var newConfig Config
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

	err = os.WriteFile(x.configPath, content, 0644)
	if err == nil {
		x.l.Debug("xray: config file saved")
	}
	return errors.WithStack(err)
}
