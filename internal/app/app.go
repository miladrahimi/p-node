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
	context    context.Context
	shutdown   chan struct{}
	Config     *config.Config
	Logger     *logger.Logger
	HttpServer *server.Server
	Xray       *xray.Xray
	Database   *database.Database
}

func New() (a *App, err error) {
	a = &App{}

	a.Config = config.New()
	if err = a.Config.Init(); err != nil {
		return nil, err
	}
	a.Logger = logger.New(a.Config.Logger.Level, a.Config.Logger.Format, a.shutdown)
	if err = a.Logger.Init(); err != nil {
		return nil, err
	}

	a.Logger.Info("app: logger and Config initialized successfully")

	a.Xray = xray.New(a.Logger, a.Config.XrayConfigPath(), a.Config.XrayBinaryPath())
	a.Database = database.New(a.Logger)
	a.HttpServer = server.New(a.Config, a.Logger, a.Xray, a.Database)

	a.Logger.Info("app: modules initialized successfully")

	a.setupSignalListener()

	return a, nil
}

func (a *App) Init() {
	a.Database.Init()
	a.Logger.Info("app: initialized successfully")
}

func (a *App) setupSignalListener() {
	var cancel context.CancelFunc
	a.context, cancel = context.WithCancel(context.Background())

	go func() {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		s := <-signalChannel
		a.Logger.Info("app: system call", zap.String("signal", s.String()))
		cancel()
	}()

	go func() {
		<-a.shutdown
		cancel()
	}()
}

func (a *App) Wait() {
	<-a.context.Done()
}

func (a *App) Shutdown() {
	a.Logger.Info("app: shutting down...")
	if a.HttpServer != nil {
		a.HttpServer.Shutdown()
	}
	if a.Xray != nil {
		a.Xray.Shutdown()
	}
	if a.Logger != nil {
		a.Logger.Shutdown()
	}
}
