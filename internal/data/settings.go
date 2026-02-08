package data

import (
	"math/rand"
	"time"

	"github.com/labstack/gommon/random"
)

// Settings is the settings struct.
type Settings struct {
	HttpPort  int    `json:"http_port" validate:"required,min=1,max=65535"` // HTTP port to serve API requests
	HttpToken string `json:"http_token" validate:"required,min=8,max=128"`  // HTTP token for authentication
}

// NewSettings creates a new settings instance.
func NewSettings(httpPort int, httpToken string) *Settings {
	return &Settings{
		HttpPort:  httpPort,
		HttpToken: httpToken,
	}
}

// DefaultSettings returns the settings with default values.
func DefaultSettings() *Settings {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return NewSettings(
		r.Intn(64536)+1000,
		random.String(16),
	)
}
