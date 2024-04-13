package app

import (
	"context"
	"github.com/miladrahimi/p-manager/pkg/logger"
	"github.com/miladrahimi/p-manager/pkg/xray"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/database"
	"github.com/miladrahimi/p-node/internal/http/server"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	Context    context.Context
	Cancel     context.CancelFunc
	Shutdown   chan struct{}
	Config     *config.Config
	Logger     *logger.Logger
	HttpServer *server.Server
	Xray       *xray.Xray
	Database   *database.Database
}

func New() (a *App, err error) {
	a = &App{}
	a.Context, a.Cancel = context.WithCancel(context.Background())
	a.Shutdown = make(chan struct{})

	a.Config = config.New()
	if err = a.Config.Init(); err != nil {
		return a, err
	}
	a.Logger = logger.New(a.Config.Logger.Level, a.Config.Logger.Format, a.Shutdown)
	if err = a.Logger.Init(); err != nil {
		return a, err
	}

	a.Xray = xray.New(a.Context, a.Logger, config.XrayConfigPath, config.XrayBinaryPath())
	a.Database = database.New(a.Logger)
	a.HttpServer = server.New(a.Config, a.Logger, a.Xray, a.Database)

	a.Logger.Info("app: constructed successfully")

	a.setupSignalListener()

	return a, nil
}

func (a *App) Init() {
	a.Database.Init()
	a.Xray.Init()
	a.Logger.Info("app: initialized successfully")
}

func (a *App) setupSignalListener() {
	go func() {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		s := <-signalChannel
		a.Logger.Info("app: signal received", zap.String("signal", s.String()))
		a.Cancel()
	}()

	go func() {
		<-a.Shutdown
		a.Cancel()
	}()
}

func (a *App) Wait() {
	a.Logger.Debug("app: waiting...")
	<-a.Context.Done()
}

func (a *App) Close() {
	a.Logger.Debug("app: closing...")
	defer a.Logger.Info("app: closed")

	if a.HttpServer != nil {
		a.HttpServer.Close()
	}
	if a.Xray != nil {
		a.Xray.Close()
	}
	if a.Logger != nil {
		a.Logger.Close()
	}
}
