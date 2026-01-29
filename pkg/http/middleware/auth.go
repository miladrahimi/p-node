package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// Authorize is a middleware that authorizes requests based on a token.
func Authorize(token func() string) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			if !authorizeToken(token(), context) {
				return echo.ErrUnauthorized
			}
			return next(context)
		}
	}
}

// authorizeToken checks if the provided token matches the token in the Authorization header.
func authorizeToken(token string, context echo.Context) bool {
	authHeader := context.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[len("Bearer "):] == token
	}
	return false
}
