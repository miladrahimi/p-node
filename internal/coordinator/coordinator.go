package coordinator

import (
	"github.com/miladrahimi/xray-manager/pkg/logger"
	"github.com/miladrahimi/xray-manager/pkg/utils"
	"github.com/miladrahimi/xray-manager/pkg/xray"
	"go.uber.org/zap"
	"xray-node/internal/config"
	"xray-node/internal/database"
)

type Coordinator struct {
	l        *logger.Logger
	config   *config.Config
	database *database.Database
	xray     *xray.Xray
}

func (c *Coordinator) Run() {
	c.initDatabase()
}

func (c *Coordinator) initDatabase() {
	var err error
	if c.database.Data.Settings.HttpPort == 1826 {
		c.database.Data.Settings.HttpPort, err = utils.FreePort()
		if err != nil {
			c.l.Fatal("coordinator: cannot find free http port", zap.Error(err))
		}
	}
	c.database.Data.Settings.XrayApiPort = c.xray.Config().ApiInbound().Port
	c.database.Save()
}

func New(c *config.Config, l *logger.Logger, d *database.Database, x *xray.Xray) *Coordinator {
	return &Coordinator{config: c, l: l, database: d, xray: x}
}
