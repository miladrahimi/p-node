package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/coordinator"
	"github.com/miladrahimi/p-node/internal/database"
	"github.com/miladrahimi/p-node/internal/http/server"
	"github.com/miladrahimi/p-node/pkg/http/client"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/xray"
	"go.uber.org/zap"
)

// App is the main application struct.
type App struct {
	context     context.Context
	cancel      context.CancelFunc
	shutdown    chan struct{}
	config      *config.Config
	logger      *logger.Logger
	httpServer  *server.Server
	httpClient  *client.Client
	xray        *xray.Xray
	coordinator *coordinator.Coordinator
	database    *database.Database
}

// New creates a new application instance.
func New() (a *App, err error) {
	a = &App{}
	a.context, a.cancel = context.WithCancel(context.Background())
	a.shutdown = make(chan struct{})

	if a.config, err = config.New(); err != nil {
		return a, errors.WithStack(err)
	}

	a.logger, err = logger.New(a.config.Logger.Level, a.config.Logger.Format, a.shutdown)
	if err != nil {
		return a, errors.WithStack(err)
	}

	a.xray = xray.New(a.context, a.logger, config.XrayLogLevel, config.XrayConfigPath, config.XrayBinaryPath())
	a.database = database.New(a.logger)
	a.httpServer = server.New(a.config, a.logger, a.xray, a.database)
	a.httpClient = client.New(a.config.HttpClient.Timeout, config.AppName, config.AppVersion)
	a.coordinator = coordinator.New(a.context, a.logger, a.config, a.database, a.httpClient, a.xray)

	a.logger.Debug("app: constructed successfully")

	a.startSignalListener()

	return a, nil
}

// Run starts the application.
func (a *App) Run() error {
	if err := a.database.Init(); err != nil {
		return errors.WithStack(err)
	}
	if err := a.xray.Init(); err != nil {
		return errors.WithStack(err)
	}
	if err := a.xray.Run(); err != nil {
		return errors.WithStack(err)
	}
	a.coordinator.Run()
	a.httpServer.Run()

	a.logger.Info("app: started successfully")
	return nil
}

// Wait waits for the application context to be canceled.
func (a *App) Wait() {
	a.logger.Debug("app: waiting...")
	<-a.context.Done()
}

// Close closes the application and its components.
func (a *App) Close() {
	a.logger.Debug("app: closing...")
	defer a.logger.Info("app: closed")

	if a.httpServer != nil {
		if err := a.httpServer.Close(); err != nil {
			a.logger.Error("http server: cannot close", zap.Error(errors.WithStack(err)))
		}
	}
	if a.xray != nil {
		if err := a.xray.Stop(); err != nil {
			a.logger.Error("xray: cannot close", zap.Error(errors.WithStack(err)))
		}
	}
	if a.logger != nil {
		a.logger.Close()
	}
}

// setupSignalListener sets up a signal listener to handle interrupt and termination signals.
func (a *App) startSignalListener() {
	go func() {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		s := <-signalChannel
		a.logger.Info("app: signal received", zap.String("signal", s.String()))
		a.cancel()
	}()

	go func() {
		<-a.shutdown
		a.cancel()
	}()
}
