package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/database"
	"github.com/miladrahimi/p-node/pkg/http/client"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/worker"
	"github.com/miladrahimi/p-node/pkg/xray"
	"go.uber.org/zap"
)

type Coordinator struct {
	l       *logger.Logger
	context context.Context
	config  *config.Config
	d       *database.Database
	xray    *xray.Xray
	client  *client.Client
}

func (c *Coordinator) Run() {
	c.l.Info("coordinator: running...")

	go worker.New("SyncWithManager", c.l, 30*time.Second, func() {
		c.l.Info("coordinator: running worker for sync...")
		if err := c.Sync(); err != nil {
			c.l.Error("coordinator: cannot sync", zap.Error(errors.WithStack(err)))
		}
	}).Start(c.context)
}

func (c *Coordinator) Sync() error {
	if c.d.Data.Manager == nil {
		return nil
	}

	remoteConfig, err := c.fetchConfig(c.d.Data.Manager)
	if err != nil {
		return errors.WithStack(err)
	}

	if !c.xray.Config().Equals(remoteConfig) {
		c.l.Info("coordinator: updating xray config...")
		c.xray.SetConfig(remoteConfig)
		go c.xray.Restart()
	}

	return nil
}

func (c *Coordinator) fetchConfig(manager *database.Manager) (*xray.Config, error) {
	url := fmt.Sprintf("%s/configs", manager.Url)
	response, err := c.client.Do("GET", url, manager.Token, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var xc xray.Config
	if err = json.Unmarshal(response, &xc); err != nil {
		return nil, errors.WithStack(err)
	}

	return &xc, nil
}

func New(
	ctx context.Context,
	l *logger.Logger,
	config *config.Config,
	d *database.Database,
	client *client.Client,
	xray *xray.Xray,
) *Coordinator {
	return &Coordinator{
		l:       l,
		config:  config,
		context: ctx,
		d:       d,
		client:  client,
		xray:    xray,
	}
}
