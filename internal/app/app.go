package app

import (
	"context"
	"github.com/miladrahimi/xray-manager/pkg/logger"
	"github.com/miladrahimi/xray-manager/pkg/xray"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"xray-node/internal/config"
	"xray-node/internal/database"
	"xray-node/internal/http/server"
)

// App integrates the modules to serve.
type App struct {
	context    context.Context
	config     *config.Config
	log        *logger.Logger
	httpServer *server.Server
	xray       *xray.Xray
	database   *database.Database
}

// New creates an instance of the application with dependencies injected.
func New() (a *App, err error) {
	a = &App{}

	a.config = config.New()
	if a.config.Init() != nil {
		return nil, err
	}
	a.log = logger.New(a.config.Logger.Level, a.config.Logger.Format)
	if a.log.Init() != nil {
		return nil, err
	}

	a.database = database.New(a.log.Engine)
	a.xray = xray.New(a.log.Engine, config.XrayConfigPath, a.config.XrayPath())
	a.httpServer = server.New(a.config, a.log.Engine, a.xray, a.database)

	a.setupSignalListener()

	return a, nil
}

// Boot initializes application modules
func (a *App) Boot() {
	a.database.Init()
	a.xray.Run()
	a.httpServer.Run()
}

// setupSignalListener sets up a listener to signals from os.
func (a *App) setupSignalListener() {
	var cancel context.CancelFunc
	a.context, cancel = context.WithCancel(context.Background())

	// Listen to SIGTERM
	go func() {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		s := <-signalChannel
		a.log.Engine.Info("app: system call", zap.String("signal", s.String()))

		cancel()
	}()

	// Listen to SIGHUP
	go func() {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, syscall.SIGHUP)

		for {
			s := <-signalChannel
			a.log.Engine.Info("app: system call", zap.String("signal", s.String()))
			a.xray.Restart()
		}
	}()
}

// Wait avoid dying app and shut it down gracefully on exit signals.
func (a *App) Wait() {
	<-a.context.Done()
}

// Shutdown closes all open resources and processes gracefully.
func (a *App) Shutdown() {
	if a.httpServer != nil {
		a.httpServer.Shutdown()
	}
	if a.xray != nil {
		a.xray.Shutdown()
	}
	if a.log != nil {
		a.log.Shutdown()
	}
}
