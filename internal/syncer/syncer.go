package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/database"
	"github.com/miladrahimi/p-node/pkg/http/client"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/worker"
	"github.com/miladrahimi/p-node/pkg/xray"
	"go.uber.org/zap"
	"time"
)

type Syncer struct {
	l        *logger.Logger
	context  context.Context
	config   *config.Config
	xray     *xray.Xray
	client   *client.Client
	database *database.Database
}

const workerInterval = 30 * time.Second

func (s *Syncer) Run() {
	s.l.Info("syncer: running...")

	go worker.New(s.context, workerInterval, func() {
		s.l.Info("syncer: running worker for sync...")
		if err := s.Sync(); err != nil {
			s.l.Error("syncer: cannot sync", zap.Error(errors.WithStack(err)))
		}
	}, func() {
		s.l.Debug("syncer: worker for sync stopped")
	}).Start()
}

func (s *Syncer) Sync() error {
	if s.database.Data.Manager == nil {
		return nil
	}

	remoteConfig, err := s.fetchConfig(s.database.Data.Manager)
	if err != nil {
		return errors.WithStack(err)
	}

	if !s.xray.Config().Equals(remoteConfig) {
		s.l.Info("syncer: updating xray config...")
		s.xray.SetConfig(remoteConfig)
		go s.xray.Restart()
	}

	return nil
}

func (s *Syncer) fetchConfig(manager *database.Manager) (*xray.Config, error) {
	url := fmt.Sprintf("%s/configs", manager.Url)
	response, err := s.client.Do("GET", url, manager.Token, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fmt.Println(string(response))

	var c xray.Config
	err = json.Unmarshal(response, &c)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &c, nil
}

func New(ctx context.Context, l *logger.Logger, config *config.Config, d *database.Database, client *client.Client, xray *xray.Xray) *Syncer {
	return &Syncer{
		l:        l,
		config:   config,
		context:  ctx,
		database: d,
		client:   client,
		xray:     xray,
	}
}
