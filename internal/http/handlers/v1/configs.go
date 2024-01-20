package v1

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/xray-manager/pkg/utils"
	"github.com/miladrahimi/xray-manager/pkg/xray"
	"net/http"
)

func ConfigsStore(x *xray.Xray) echo.HandlerFunc {
	return func(c echo.Context) error {
		var config xray.Config
		if err := c.Bind(&config); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		if err := config.Validate(); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		isShadowsocksPortNew := false
		if config.ShadowsocksInboundIndex() != -1 {
			if x.Config().ShadowsocksInboundIndex() == -1 {
				isShadowsocksPortNew = true
			} else if config.ShadowsocksInbound().Port != x.Config().ShadowsocksInbound().Port {
				isShadowsocksPortNew = true
			}
		}
		if isShadowsocksPortNew && !utils.PortFree(config.ShadowsocksInbound().Port) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": fmt.Sprintf("Shadowsocks port is already in use."),
			})
		}

		x.Config().Locker.Lock()
		defer x.Config().Locker.Unlock()

		config.UpdateApiInbound(x.Config().ApiInbound().Port)
		x.SetConfig(&config)
		x.Restart()

		return c.JSON(http.StatusOK, map[string]string{
			"message": "The configs stored successfully.",
		})
	}
}
