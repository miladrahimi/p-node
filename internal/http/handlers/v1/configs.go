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
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		if err := config.Validate(); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		if config.DirectInbound() != nil {
			p := config.DirectInbound().Port
			if x.Config().DirectInbound() == nil || p != x.Config().DirectInbound().Port {
				if !utils.PortFree(p) {
					return c.JSON(http.StatusUnprocessableEntity, map[string]string{
						"message": fmt.Sprintf("The direct inbound port '%d' is already in use", p),
					})
				}
			}
		}

		x.SetConfig(&config)
		go x.Restart()

		return c.JSON(http.StatusOK, map[string]string{
			"message": "The configs stored successfully.",
		})
	}
}
