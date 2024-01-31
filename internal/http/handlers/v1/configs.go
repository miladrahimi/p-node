package v1

import (
	"fmt"
	"github.com/labstack/echo/v4"
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

		config.ApiInbound().Port = x.Config().ApiInbound().Port
		x.SetConfig(&config)
		go x.Restart()

		return c.JSON(http.StatusOK, map[string]string{
			"message": "The configs stored successfully.",
		})
	}
}
