package xray

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/p-node/pkg/util"
	"github.com/miladrahimi/p-node/pkg/xray"
	"github.com/miladrahimi/p-node/pkg/xray/config"
)

// ConfigStore stores/updates the Xray configuration.
func ConfigStore(x *xray.Xray) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		var xc config.Config
		if err = ctx.Bind(&xc); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		if err = xc.Validate(); err != nil {
			return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		for _, i := range xc.Inbounds {
			isFree := util.PortFree(i.Port)
			if i.Tag != "api" && i.Tag != "remote" && !isFree {
				return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
					"message": fmt.Sprintf("The port '%s.%d' is already in use", i.Tag, i.Port),
				})
			}
			if i.Tag == "remote" && !isFree {
				currentInbound := x.Config().FindInbound("remote")
				if currentInbound == nil || currentInbound.Port != i.Port {
					return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
						"message": fmt.Sprintf("The port '%s.%d' is already in use", i.Tag, i.Port),
					})
				}
			}
			if i.Tag == "api" {
				if i.Port, err = util.FreePort(); err != nil {
					return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
						"message": fmt.Sprintf("Cannot generate API inbound port, err: %v", err.Error()),
					})
				}
			}
		}

		if err = x.Reconfigure(&xc); err != nil {
			return errors.WithStack(err)
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "The configs stored successfully.",
		})
	}
}
