package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func HomeShow() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "I am fine!",
		})
	}
}
