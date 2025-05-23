package xray

import (
	"context"
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/utils"
	"github.com/miladrahimi/p-node/pkg/logger"
	stats "github.com/xtls/xray-core/app/stats/command"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

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

func (x *Xray) saveConfig() error {
	x.l.Debug("xray: saving config file...")

	content, err := json.Marshal(x.config)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.WriteFile(x.configPath, content, 0755)
	if err == nil {
		x.l.Debug("xray: config file saved")
	}
	return errors.WithStack(err)
}

func (x *Xray) Run() error {
	x.l.Debug("xray: running...")

	x.locker.Lock()
	defer x.locker.Unlock()

	if err := x.saveConfig(); err != nil {
		return errors.WithStack(err)
	}

	go x.runCore()

	err := x.connect()
	return errors.WithStack(err)
}

func (x *Xray) Init() error {
	err := x.loadConfig()
	return errors.WithStack(err)
}

func (x *Xray) runCore() {
	x.l.Debug("xray: running core...")

	if !utils.FileExist(x.binaryPath) {
		x.l.Fatal("xray: binary not found", zap.String("path", x.binaryPath))
	}

	x.command = exec.Command(x.binaryPath, "-c", x.configPath)
	x.command.Stderr = os.Stderr
	x.command.Stdout = os.Stdout

	x.l.Info("xray: executing the binary...", zap.String("path", x.binaryPath))
	if err := x.command.Run(); err != nil && err.Error() != "signal: killed" {
		x.l.Fatal("xray: cannot execute the binary", zap.Error(errors.WithStack(err)))
	}
}

func (x *Xray) Restart() {
	x.l.Info("xray: restarting...")

	if err := x.Close(); err != nil {
		x.l.Error("xray: cannot close", zap.Error(errors.WithStack(err)))
	}

	if err := x.Run(); err != nil {
		x.l.Fatal("xray: cannot run again", zap.Error(errors.WithStack(err)))
	}
}

func (x *Xray) Close() error {
	x.l.Debug("xray: closing...")

	x.locker.Lock()
	defer x.locker.Unlock()

	if x.connection != nil {
		x.l.Debug("xray: disconnecting the api connection...")
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

func (x *Xray) connect() error {
	x.l.Debug("xray: connecting to api...")

	inbound := x.config.FindInbound("api")
	if inbound == nil {
		return errors.New("no inbound inbound")
	}

	c, cancel := context.WithTimeout(x.context, 10*time.Second)
	defer cancel()

	address := "127.0.0.1:" + strconv.Itoa(inbound.Port)
	var err error

	for {
		select {
		case <-c.Done():
			return errors.New("connection to Xray api timed out")
		default:
			time.Sleep(time.Second)
			x.connection, err = grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				x.l.Debug("xray: trying to connect to api", zap.Error(errors.WithStack(err)))
			} else {
				x.l.Debug("xray: connected to api successfully")
				return nil
			}
		}
	}
}

func (x *Xray) Config() *Config {
	return x.config
}

func (x *Xray) SetConfig(config *Config) {
	x.config = config
}

func (x *Xray) QueryStats() ([]*stats.Stat, error) {
	client := stats.NewStatsServiceClient(x.connection)
	qs, err := client.QueryStats(context.Background(), &stats.QueryStatsRequest{Reset_: true})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return qs.GetStat(), nil
}

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
