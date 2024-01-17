package v1

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"xray-node/internal/xray"
)

func StatsShow(x *xray.Xray) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, x.QueryStats())
	}
}
