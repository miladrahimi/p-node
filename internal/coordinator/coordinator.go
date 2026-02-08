package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/data"
	"github.com/miladrahimi/p-node/pkg/database"
	httpClient "github.com/miladrahimi/p-node/pkg/http/client"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/worker"
	"github.com/miladrahimi/p-node/pkg/xray"
	config2 "github.com/miladrahimi/p-node/pkg/xray/config"
)

// Coordinator represents the app coordinator which manages xray and database.
type Coordinator struct {
	logger     *logger.Logger
	context    context.Context
	config     *config.Config
	data       *database.Database[data.Data]
	xray       *xray.Xray
	httpClient *httpClient.Client
}

// New creates a new coordinator.
func New(
	ctx context.Context,
	l *logger.Logger,
	c *config.Config,
	d *database.Database[data.Data],
	hc *httpClient.Client,
	x *xray.Xray,
) *Coordinator {
	return &Coordinator{
		logger:     l,
		config:     c,
		context:    ctx,
		data:       d,
		httpClient: hc,
		xray:       x,
	}
}

// Run starts the coordinator.
func (c *Coordinator) Run() {
	c.logger.Info("coordinator: running...")

	go worker.New("SyncWithManager", c.logger, 30*time.Second, func() error {
		return c.Sync()
	}).Start(c.context)
}

// Sync syncs the xray config with the associated P-Manager if it exists.
func (c *Coordinator) Sync() error {
	manager := c.data.Data().Manager
	if manager == nil {
		return nil
	}

	managerXrayConfig, err := c.fetchConfig(manager)
	if err != nil {
		return errors.WithStack(err)
	}

	localXrayConfig := c.xray.Config()
	if localXrayConfig.Equals(managerXrayConfig) {
		return nil
	}

	c.logger.Info("coordinator: reconfiguring xray...")

	return errors.WithStack(c.xray.Reconfigure(managerXrayConfig))
}

// fetchConfig fetches the xray config from the given P-Manager.
func (c *Coordinator) fetchConfig(manager *data.Manager) (*config2.Config, error) {
	url := fmt.Sprintf("%s/configs", manager.Url)
	response, err := c.httpClient.Do("GET", url, manager.Token, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var xc config2.Config
	err = json.Unmarshal(response, &xc)
	return &xc, errors.WithStack(err)
}
