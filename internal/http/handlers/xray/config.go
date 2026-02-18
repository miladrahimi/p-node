package xray

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/p-node/pkg/util"
	"github.com/miladrahimi/p-node/pkg/xray"
	"github.com/miladrahimi/p-node/pkg/xray/config"
	"github.com/miladrahimi/p-node/pkg/xray/config/component"
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

		currentPorts := map[int]int{}
		for _, inbound := range x.Config().Inbounds {
			if inbound.Tag == "api" {
				continue
			}
			currentPorts[inbound.Port]++
		}

		usedPorts := map[int]bool{}
		var apiInbound *component.Inbound
		for _, i := range xc.Inbounds {
			if i.Tag == "api" {
				if apiInbound != nil {
					return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
						"message": "Only one API inbound is allowed.",
					})
				}
				apiInbound = i
				continue
			}

			if usedPorts[i.Port] {
				return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
					"message": fmt.Sprintf("The port '%d' is used by multiple inbounds.", i.Port),
				})
			}
			usedPorts[i.Port] = true

			isFree := util.PortFree(i.Port)
			if !isFree {
				if currentPorts[i.Port] == 0 {
					return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
						"message": fmt.Sprintf("The port '%d' for '%s' is already in use", i.Port, i.Tag),
					})
				}
				currentPorts[i.Port]--
			}
		}

		if apiInbound != nil {
			if apiInbound.Port, err = util.FreePort(); err != nil {
				return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
					"message": fmt.Sprintf("Cannot generate API inbound port, err: %v", err.Error()),
				})
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
