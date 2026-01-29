package v1

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/miladrahimi/p-node/internal/database"
)

type ManagerStoreRequest struct {
	Url   string `json:"url" validate:"omitempty,url,min=1,max=1024"`
	Token string `json:"token" validate:"omitempty,min=1,max=128"`
}

// ManagerStore stores the associated P-Manager config.
func ManagerStore(d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r ManagerStoreRequest
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		if err := c.Validate(&r); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		if r.Url == "" {
			d.Data.Manager = nil
		} else {
			d.Data.Manager = &database.Manager{
				Url:   r.Url,
				Token: r.Token,
			}
		}

		if err := d.Save(); err != nil {
			return errors.WithStack(err)
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"manager": r,
		})
	}
}
