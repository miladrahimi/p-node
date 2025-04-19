package app

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/database"
	"github.com/miladrahimi/p-node/internal/http/server"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/xray"
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

	a.Xray = xray.New(a.Context, a.Logger, config.XrayLogLevel, config.XrayConfigPath, config.XrayBinaryPath())
	a.Database = database.New(a.Logger)
	a.HttpServer = server.New(a.Config, a.Logger, a.Xray, a.Database)

	a.Logger.Debug("app: constructed successfully")

	a.startSignalListener()

	return a, nil
}

func (a *App) Start() error {
	if err := a.Database.Init(); err != nil {
		return errors.WithStack(err)
	}
	if err := a.Xray.Init(); err != nil {
		return errors.WithStack(err)
	}
	if err := a.Xray.Run(); err != nil {
		return errors.WithStack(err)
	}
	a.HttpServer.Run()

	a.Logger.Info("app: started successfully")
	return nil
}

func (a *App) startSignalListener() {
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
		if err := a.Xray.Close(); err != nil {
			a.Logger.Error("xray: cannot close", zap.Error(errors.WithStack(err)))
		}
	}
	if a.Logger != nil {
		a.Logger.Close()
	}
}
