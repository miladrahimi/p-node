package v1

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/xray-manager/pkg/xray"
	"net/http"
)

func Action(x *xray.Xray) echo.HandlerFunc {
	return func(c echo.Context) error {
		var config xray.Config
		if err := c.Bind(&config); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		if err := validator.New().Struct(config); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		stats := x.QueryStats()
		config.UpdateApiInbound(x.Config().ApiInbound().Port)
		x.SetConfig(&config)
		go x.Restart()

		return c.JSON(http.StatusOK, stats)
	}
}