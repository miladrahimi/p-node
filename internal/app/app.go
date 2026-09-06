package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/coordinator"
	"github.com/miladrahimi/p-node/internal/data"
	"github.com/miladrahimi/p-node/internal/http/server"
	"github.com/miladrahimi/p-node/pkg/database"
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
	database    *database.Database[data.Data]
}

// New creates a new application instance.
func New() (a *App, err error) {
	a = &App{}
	a.context, a.cancel = context.WithCancel(context.Background())
	a.shutdown = make(chan struct{})

	root, err := os.Getwd()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if a.config, err = config.New(root); err != nil {
		return a, errors.WithStack(err)
	}
	c := a.config

	if a.logger, err = logger.New(c.Logger.Level, c.Logger.Format, a.shutdown); err != nil {
		return a, errors.WithStack(err)
	}
	l := a.logger

	if a.database, err = database.New(config.DatabaseDirectory(root), data.Default()); err != nil {
		return a, errors.WithStack(err)
	}

	a.xray = xray.New(a.context, l, c.Xray.LogLevel, config.XrayConfigPath(root), config.XrayBinaryPath(root))
	a.httpClient = client.New(c.HttpClient.Timeout, config.AppName, config.AppVersion)
	a.coordinator = coordinator.New(a.context, l, c, a.database, a.httpClient, a.xray)
	a.httpServer = server.New(c, l, a.xray, a.coordinator, a.database)

	l.Debug("app: constructed successfully")

	a.startSignalListener()

	return a, nil
}

// Run starts the application.
func (a *App) Run() error {
	if err := a.database.Init(); err != nil {
		return errors.WithStack(err)
	}
	if p := a.config.HttpServer.Port; p != 0 && p != a.database.Data().Settings.HttpPort {
		a.database.Data().Settings.HttpPort = p
		if err := a.database.Save(); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := a.xray.Load(); err != nil {
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
			a.logger.Error("cannot close http server", zap.Error(errors.WithStack(err)))
		}
	}
	if a.xray != nil {
		if err := a.xray.Stop(); err != nil {
			a.logger.Error("cannot close xray", zap.Error(errors.WithStack(err)))
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
