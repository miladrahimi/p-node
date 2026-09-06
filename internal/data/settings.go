package data

import (
	"math/rand"
	"time"

	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/p-node/pkg/util"
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
	return NewSettings(randomFreePort(), random.String(16))
}

// randomFreePort picks a random unprivileged port that is currently free.
// It falls back to the last candidate if none of the attempts is free.
func randomFreePort() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 0
	for i := 0; i < 16; i++ {
		port = r.Intn(64512) + 1024
		if util.PortFree(port) {
			break
		}
	}
	return port
}
