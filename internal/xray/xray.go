package xray

import (
	"context"
	stats "github.com/xtls/xray-core/app/stats/command"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"xray-node/internal/config"
	"xray-node/internal/database"
	"xray-node/internal/utils"
)

var configPath = "storage/xray.json"
var configTemplatePath = "configs/xray.json"
var binaryPaths = map[string]string{
	"darwin": "third_party/xray-macos-arm64/xray",
	"linux":  "third_party/xray-linux-64/xray",
}

type Xray struct {
	command    *exec.Cmd
	log        *zap.Logger
	connection *grpc.ClientConn
	lock       sync.Mutex
	config     *config.Config
	database   *database.Database
}

// binaryPath returns the path of Xray core binary for current OS.
func (x *Xray) binaryPath() string {
	if path, found := binaryPaths[runtime.GOOS]; found {
		return path
	}
	return binaryPaths["linux"]
}

// initConfig stores init config file if there is no config file.
func (x *Xray) initConfig() {
	if !utils.FileExist(configPath) {
		templateContent, err := os.ReadFile(configTemplatePath)
		if err != nil {
			x.log.Fatal("xray: cannot open config template file", zap.Error(err))
		}

		apiPort := strconv.Itoa(x.database.Data.Settings.InternalPort)
		content := strings.ReplaceAll(string(templateContent), "2401", apiPort)

		if err = os.WriteFile(configPath, []byte(content), 0644); err != nil {
			x.log.Fatal("xray: cannot save init config file", zap.Error(err))
		}
	}
}

// Run prepare and starts the Xray core process.
func (x *Xray) Run() {
	x.initConfig()
	go x.runCore()
	x.connectGrpc()
}

// runCore runs Xray core.
func (x *Xray) runCore() {
	x.command = exec.Command(x.binaryPath(), "-c", configPath)
	x.command.Stderr = os.Stderr
	x.command.Stdout = os.Stdout

	x.log.Debug("xray: starting the xray core...")
	if err := x.command.Run(); err != nil && err.Error() != "signal: killed" {
		x.log.Fatal("xray: cannot start the xray core", zap.Error(err))
	}
}

// Restart closes and runs the Xray core.
func (x *Xray) Restart() {
	x.log.Info("xray: restarting the xray core...")
	x.Shutdown()
	x.Run()
}

// Shutdown closes Xray core process.
func (x *Xray) Shutdown() {
	x.log.Debug("xray: shutting down the xray core...")
	if x.connection != nil {
		_ = x.connection.Close()
	}
	if x.command.Process != nil {
		if err := x.command.Process.Kill(); err != nil {
			x.log.Error("xray: failed to shutdown the xray core", zap.Error(err))
		} else {
			x.log.Debug("xray: the xray core stopped successfully")
		}
	} else {
		x.log.Debug("xray: the xray core is already stopped")
	}
}

// connectGrpc connects to the GRPC APIs provided by Xray core.
func (x *Xray) connectGrpc() {
	x.log.Debug("xray: connecting to xray core grpc...")

	address := "127.0.0.1:" + strconv.Itoa(x.database.Data.Settings.InternalPort)
	var err error

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		x.connection, err = grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			x.log.Debug("xray: cannot connect the xray core grpc", zap.Int("try", i))
		} else {
			x.log.Debug("xray: connected to grpc server successfully.")
			return
		}
	}

	x.log.Fatal("xray: cannot connect the xray core grpc", zap.Error(err))
}

// QueryStats fetches the traffic stats from Xray core.
func (x *Xray) QueryStats() []*stats.Stat {
	client := stats.NewStatsServiceClient(x.connection)
	qs, err := client.QueryStats(context.Background(), &stats.QueryStatsRequest{Reset_: true})
	if err != nil {
		x.log.Error("xray: cannot query stats", zap.Error(err))
	}
	return qs.GetStat()
}

func (x *Xray) LoadConfigs() string {
	content, err := os.ReadFile(configPath)
	if err != nil {
		x.log.Fatal("xray: cannot load config file", zap.Error(err))
	}
	return string(content)
}

func (x *Xray) SaveConfigs(data []byte) {
	apiPort := strconv.Itoa(x.database.Data.Settings.InternalPort)
	content := strings.ReplaceAll(string(data), "2401", apiPort)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		x.log.Fatal("xray: cannot save init config file", zap.Error(err))
	}
}

// New creates a new instance of Xray.
func New(l *zap.Logger, d *database.Database, c *config.Config) *Xray {
	return &Xray{log: l, config: c, database: d}
}
