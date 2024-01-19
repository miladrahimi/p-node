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
	"xray-node/internal/coordinator"
	"xray-node/internal/database"
	"xray-node/internal/http/server"
)

type App struct {
	context     context.Context
	config      *config.Config
	log         *logger.Logger
	httpServer  *server.Server
	xray        *xray.Xray
	database    *database.Database
	coordinator *coordinator.Coordinator
}

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

	a.xray = xray.New(a.log.Engine, a.config.XrayConfigPath(), a.config.XrayBinaryPath())
	a.database = database.New(a.log.Engine)
	a.coordinator = coordinator.New(a.config, a.log.Engine, a.database, a.xray)
	a.httpServer = server.New(a.config, a.log.Engine, a.xray, a.database)

	a.setupSignalListener()

	return a, nil
}

func (a *App) Boot() {
	a.xray.Run()
	a.database.Init()
	a.coordinator.Run()
	a.httpServer.Run()
}

func (a *App) setupSignalListener() {
	var cancel context.CancelFunc
	a.context, cancel = context.WithCancel(context.Background())

	go func() {
		signalChannel := make(chan os.Signal, 2)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		s := <-signalChannel
		a.log.Engine.Info("app: system call", zap.String("signal", s.String()))

		cancel()
	}()

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

func (a *App) Wait() {
	<-a.context.Done()
}

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
