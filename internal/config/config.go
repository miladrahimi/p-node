package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/miladrahimi/p-node/pkg/util"
)

const AppName = "P-Node"
const AppVersion = "v26.9.6"

// Config represents the application configuration.
type Config struct {
	HttpClient struct {
		Timeout int `json:"timeout" validate:"required,min=10,max=60000"`
	} `json:"http_client" validate:"required"`

	Xray struct {
		LogLevel string `json:"log_level" validate:"required,oneof=debug info warning error none"`
	} `json:"xray" validate:"required"`

	Logger struct {
		Level  string `json:"level" validate:"required,oneof=debug info warn error"`
		Format string `json:"format" validate:"required,oneof='2006-01-02 15:04:05.000'"`
	} `json:"logger" validate:"required"`
}

// New creates a new instance of Config and loads the default and local config files.
func New(root string) (*Config, error) {
	c := &Config{}

	content, err := os.ReadFile(defaultConfigPath(root))
	if err != nil {
		return c, errors.WithStack(err)
	}
	err = json.Unmarshal(content, c)
	if err != nil {
		return c, errors.WithStack(err)
	}

	lcp := localConfigPath(root)
	if util.FileExist(lcp) {
		content, err = os.ReadFile(lcp)
		if err != nil {
			return c, errors.WithStack(err)
		}
		if err = json.Unmarshal(content, c); err != nil {
			return c, errors.WithStack(err)
		}

		var contentBytes []byte
		contentBytes, err = json.MarshalIndent(c, "", "  ")
		if err != nil {
			return c, errors.WithStack(err)
		}
		if err = os.WriteFile(lcp, contentBytes, 0644); err != nil {
			return c, errors.WithStack(err)
		}
	}

	fmt.Println("Config:", c.String())

	return c, errors.WithStack(validator.New().Struct(c))
}

// String returns a string representation of the configuration.
func (c *Config) String() string {
	j, err := json.Marshal(c)
	if err != nil {
		return err.Error()
	}
	return string(j)
}
