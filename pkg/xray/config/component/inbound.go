package component

type Inbound struct {
	Listen   string `json:"listen" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Protocol string `json:"protocol" validate:"required"`
	Settings any    `json:"settings" validate:"required"`
	Tag      string `json:"tag" validate:"required"`
}

type InboundSettings struct {
	Address  string `json:"address,omitempty"`
	Clients  []any  `json:"clients,omitempty" validate:"omitempty,dive"`
	Network  string `json:"network,omitempty"`
	Method   string `json:"method,omitempty"`
	Password string `json:"password,omitempty"`
}
