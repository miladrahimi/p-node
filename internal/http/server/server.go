package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/miladrahimi/xray-manager/pkg/http/middleware"
	"github.com/miladrahimi/xray-manager/pkg/http/validator"
	"github.com/miladrahimi/xray-manager/pkg/xray"
	"go.uber.org/zap"
	"net/http"
	"time"
	"xray-node/internal/config"
	"xray-node/internal/database"
	"xray-node/internal/http/handlers"
	"xray-node/internal/http/handlers/v1"
)

type Server struct {
	Engine   *echo.Echo
	config   *config.Config
	log      *zap.Logger
	xray     *xray.Xray
	database *database.Database
}

// Run defines the required HTTP routes and starts the HTTP Server.
func (s *Server) Run() {
	s.Engine.Use(echoMiddleware.CORS())
	s.Engine.Use(middleware.Logger(s.log))

	s.Engine.GET("/", handlers.HomeShow())

	g2 := s.Engine.Group("/v1")
	g2.Use(middleware.Authorize(s.database.Data.Settings.HttpToken))

	g2.GET("/action", v1.Action(s.xray))

	go func() {
		address := fmt.Sprintf("0.0.0.0:%d", s.database.Data.Settings.HttpPort)
		if err := s.Engine.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Fatal("http server: failed to start", zap.String("address", address), zap.Error(err))
		}
	}()
}

// Shutdown closes the HTTP Server.
func (s *Server) Shutdown() {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Engine.Shutdown(c); err != nil {
		s.log.Error("http server: failed to close", zap.Error(err))
	} else {
		s.log.Debug("http server: closed successfully")
	}
}

// New creates a new instance of HTTP Server.
func New(config *config.Config, l *zap.Logger, x *xray.Xray, d *database.Database) *Server {
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.New()

	return &Server{Engine: e, config: config, log: l, xray: x, database: d}
}
