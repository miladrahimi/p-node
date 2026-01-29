package validator

import (
	"net/http"

	pg "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Validator is a struct that holds the validator instance.
type Validator struct {
	validator *pg.Validate
}

// New creates a new instance of Validator.
func New() *Validator {
	return &Validator{validator: pg.New()}
}

// Validate validates the given interface using the validator instance.
func (cv *Validator) Validate(i interface{}) error {
	v := pg.New(pg.WithRequiredStructEnabled())
	if err := v.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
