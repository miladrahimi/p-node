package component

type Api struct {
	Tag      string   `json:"tag" validate:"required"`
	Services []string `json:"services" validate:"required"`
}
