package v1

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/p-manager/pkg/utils"
	"github.com/miladrahimi/p-manager/pkg/xray"
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
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		if c.Request().Header.Get("X-App-Name") != "P-Manager" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": fmt.Sprintf("Unknown client."),
			})
		}

		for _, i := range config.Inbounds {
			if i.Tag != "api" && !utils.PortFree(i.Port) {
				return c.JSON(http.StatusUnprocessableEntity, map[string]string{
					"message": fmt.Sprintf("The port '%s.%d' is already in use", i.Tag, i.Port),
				})
			}
		}

		x.SetConfig(&config)

		go x.Restart()

		return c.JSON(http.StatusOK, map[string]string{
			"message": "The configs stored successfully.",
		})
	}
}
