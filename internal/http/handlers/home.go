package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HomeShow returns the home page.
func HomeShow() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "I am fine!",
		})
	}
}
