package v1

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"xray-node/internal/xray"
)

func ConfigsShow(x *xray.Xray) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.String(http.StatusOK, x.LoadConfigs())
	}
}

func ConfigsStore(x *xray.Xray) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}

		x.SaveConfigs(body)
		go x.Restart()

		return c.JSON(http.StatusOK, map[string]string{
			"message": "success",
		})
	}
}
