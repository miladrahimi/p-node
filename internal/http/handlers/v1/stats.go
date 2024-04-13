package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/p-node/pkg/xray"
	"net/http"
)

func StatsShow(x *xray.Xray) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, x.QueryStats())
	}
}
