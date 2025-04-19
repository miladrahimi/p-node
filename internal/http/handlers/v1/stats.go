package v1

import (
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/p-node/pkg/xray"
	"net/http"
)

func StatsShow(x *xray.Xray) echo.HandlerFunc {
	return func(c echo.Context) error {
		stats, err := x.QueryStats()
		if err != nil {
			return errors.WithStack(err)
		}
		return c.JSON(http.StatusOK, stats)
	}
}
