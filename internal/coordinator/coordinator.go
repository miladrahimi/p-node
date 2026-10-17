package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/miladrahimi/p-node/internal/config"
	"github.com/miladrahimi/p-node/internal/data"
	"github.com/miladrahimi/p-node/pkg/database"
	httpClient "github.com/miladrahimi/p-node/pkg/http/client"
	"github.com/miladrahimi/p-node/pkg/logger"
	"github.com/miladrahimi/p-node/pkg/paqet"
	"github.com/miladrahimi/p-node/pkg/worker"
	"github.com/miladrahimi/p-node/pkg/xray"
	xrayConfig "github.com/miladrahimi/p-node/pkg/xray/config"
	"github.com/miladrahimi/p-node/pkg/xray/config/component"
	"go.uber.org/zap"
)

// Coordinator represents the app coordinator which manages xray, paqet and database.
type Coordinator struct {
	logger     *logger.Logger
	context    context.Context
	config     *config.Config
	db         *database.Database[data.Data]
	xray       *xray.Xray
	paqet      *paqet.Paqet
	httpClient *httpClient.Client

	networkMu sync.Mutex
	network   *paqet.Network
}

// New creates a new coordinator.
func New(
	ctx context.Context,
	l *logger.Logger,
	c *config.Config,
	d *database.Database[data.Data],
	hc *httpClient.Client,
	x *xray.Xray,
	p *paqet.Paqet,
) *Coordinator {
	return &Coordinator{
		logger:     l,
		config:     c,
		context:    ctx,
		db:         d,
		httpClient: hc,
		xray:       x,
		paqet:      p,
	}
}

// Run starts the coordinator.
func (c *Coordinator) Run() {
	c.logger.Info("coordinator: running...")

	worker.New("SyncWithManager", c.logger, 30*time.Second, func() error {
		return c.Sync()
	}).Start(c.context)
}

// Sync syncs the xray config with the associated P-Manager if it exists.
func (c *Coordinator) Sync() error {
	d := c.db.Data()
	manager := d.Manager
	if manager == nil {
		return nil
	}

	xrayConfigFromManager, err := c.fetchConfig(manager)
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(c.Apply(xrayConfigFromManager))
}

// Apply applies a config from P-Manager: it reconfigures Xray (a no-op when
// the config matches the running one) and the paqet server it carries.
func (c *Coordinator) Apply(xc *xrayConfig.Config) error {
	if err := c.xray.Reconfigure(xc); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(c.applyPaqet(xc.Paqet))
}

// ApplyStored starts the paqet server of the stored config (on boot).
func (c *Coordinator) ApplyStored() error {
	return errors.WithStack(c.applyPaqet(c.xray.Config().Paqet))
}

// Close stops the paqet server.
func (c *Coordinator) Close() error {
	return errors.WithStack(c.paqet.Stop())
}

// applyPaqet runs the paqet server for p, or stops it when p is nil.
func (c *Coordinator) applyPaqet(p *component.Paqet) error {
	if p == nil {
		return errors.WithStack(c.paqet.Reconfigure(nil))
	}
	network, err := c.paqetNetwork()
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(c.paqet.Reconfigure(paqet.ServerConfig(*network, p.Port, p.Key)))
}

// paqetNetwork returns the network paqet uses: the configured override, or the
// detected default route (detected once and cached).
func (c *Coordinator) paqetNetwork() (*paqet.Network, error) {
	c.networkMu.Lock()
	defer c.networkMu.Unlock()

	if c.network != nil {
		return c.network, nil
	}
	if o := c.config.Paqet; o.Set() {
		c.network = &paqet.Network{Interface: o.Interface, Address: o.Address, RouterMac: o.RouterMac}
		return c.network, nil
	}
	network, err := paqet.DetectNetwork()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c.logger.Info("coordinator: paqet network detected", zap.String("network", network.String()))
	c.network = network
	return network, nil
}

// fetchConfig fetches the xray config from the given P-Manager.
func (c *Coordinator) fetchConfig(manager *data.Manager) (*xrayConfig.Config, error) {
	url := fmt.Sprintf("%s/config", manager.Url)
	response, err := c.httpClient.Do("GET", url, manager.Token, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var xc xrayConfig.Config
	err = json.Unmarshal(response, &xc)
	return &xc, errors.WithStack(err)
}
