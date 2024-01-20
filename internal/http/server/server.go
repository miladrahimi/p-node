package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/miladrahimi/xray-manager/pkg/logger"
	"github.com/miladrahimi/xray-manager/pkg/routing/middleware"
	"github.com/miladrahimi/xray-manager/pkg/routing/validator"
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
	engine   *echo.Echo
	config   *config.Config
	xray     *xray.Xray
	database *database.Database
	l        *logger.Logger
}

// Run defines the required HTTP routes and starts the HTTP Server.
func (s *Server) Run() {
	s.engine.Use(echoMiddleware.CORS())
	s.engine.Use(middleware.Logger(s.l))

	s.engine.GET("/", handlers.HomeShow())

	g2 := s.engine.Group("/v1")
	g2.Use(middleware.Authorize(s.database.Data.Settings.HttpToken))

	g2.GET("/stats", v1.StatsShow(s.xray))
	g2.POST("/configs", v1.ConfigsStore(s.xray))

	go func() {
		address := fmt.Sprintf("%s:%d", "0.0.0.0", s.database.Data.Settings.HttpPort)
		if err := s.engine.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.l.Fatal("http server: cannot start", zap.String("address", address), zap.Error(err))
		}
	}()
}

// Shutdown closes the HTTP Server.
func (s *Server) Shutdown() {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.engine.Shutdown(c); err != nil {
		s.l.Error("http server: cannot close", zap.Error(err))
	} else {
		s.l.Debug("http server: closed successfully")
	}
}

// New creates a new instance of HTTP Server.
func New(config *config.Config, l *logger.Logger, x *xray.Xray, d *database.Database) *Server {
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.New()

	return &Server{engine: e, config: config, l: l, xray: x, database: d}
}
