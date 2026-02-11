package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/coordinator"
	"github.com/miladrahimi/p-node/internal/data"
	rootHandler "github.com/miladrahimi/p-node/internal/http/handlers"
	xrayHandler "github.com/miladrahimi/p-node/internal/http/handlers/xray"
	"github.com/miladrahimi/p-node/pkg/database"
	cm "github.com/miladrahimi/p-node/pkg/http/middleware"
	"github.com/miladrahimi/p-node/pkg/http/validator"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/xray"
	"go.uber.org/zap"
)

// Server represents an HTTP server.
type Server struct {
	engine      *echo.Echo
	config      *config.Config
	logger      *logger.Logger
	xray        *xray.Xray
	coordinator *coordinator.Coordinator
	db          *database.Database[data.Data]
}

// New creates a new instance of HTTP Server.
func New(
	c *config.Config,
	l *logger.Logger,
	x *xray.Xray,
	cdr *coordinator.Coordinator,
	db *database.Database[data.Data],
) *Server {
	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.New()

	return &Server{engine: e, config: c, logger: l, xray: x, coordinator: cdr, db: db}
}

// Run defines the required HTTP routes and starts the HTTP Server.
func (s *Server) Run() {
	s.engine.Use(em.CORS())
	s.engine.Use(cm.Logger(s.logger))
	s.engine.Use(cm.General())

	// Guest APIs
	g1 := s.engine.Group("")
	g1.GET("/", rootHandler.HomeShow())

	// Authenticated APIs
	g2 := s.engine.Group("")
	g2.Use(cm.Authorize(func() string { return s.db.Data().Settings.HttpToken }))
	g2.POST("/manager", rootHandler.ManagerStore(s.db, s.coordinator))
	g2.GET("/xray/stats", xrayHandler.StatsShow(s.xray))
	g2.POST("/xray/config", xrayHandler.ConfigStore(s.xray))

	go func() {
		address := fmt.Sprintf("%s:%d", "0.0.0.0", s.db.Data().Settings.HttpPort)
		if err := s.engine.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("http server: cannot start", zap.String("address", address), zap.Error(err))
		}
	}()
}

// Close closes the HTTP Server.
func (s *Server) Close() error {
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.engine.Shutdown(c); err != nil {
		return errors.WithStack(err)
	}

	s.logger.Debug("http server: closed successfully")
	return nil
}
