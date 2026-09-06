package xray

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/p-node/pkg/xray"
	"github.com/miladrahimi/p-node/pkg/xray/config"
)

// ConfigStore stores/updates the Xray configuration.
func ConfigStore(x *xray.Xray) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var xc config.Config
		if err := ctx.Bind(&xc); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		if err := xc.Validate(); err != nil {
			return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		if err := x.Reconfigure(&xc); err != nil {
			if errors.Is(err, xray.ErrPortConflict) {
				return ctx.JSON(http.StatusUnprocessableEntity, map[string]string{
					"message": err.Error(),
				})
			}
			return errors.WithStack(err)
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "The configs stored successfully.",
		})
	}
}
