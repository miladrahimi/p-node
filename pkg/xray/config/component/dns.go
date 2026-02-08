package component

type Dns struct {
	Servers []string `json:"servers" validate:"required"`
}
