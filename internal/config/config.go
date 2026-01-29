package config

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/miladrahimi/p-node/internal/utils"
)

const AppName = "P-Node"
const AppVersion = "v25.7.20"

const XrayConfigPath = "storage/app/xray.json"
const XrayLogLevel = "debug"

const HttpTimeout = 20

const defaultConfigPath = "configs/main.defaults.json"
const localConfigPath = "configs/main.json"

var xrayBinaryPaths = map[string]string{
	"darwin": "third_party/xray-macos-arm64/xray",
	"linux":  "third_party/xray-linux-64/xray",
}

func XrayBinaryPath() string {
	if path, found := xrayBinaryPaths[runtime.GOOS]; found {
		return path
	}
	return xrayBinaryPaths["linux"]
}

type Config struct {
	HttpClient struct {
		Timeout int `json:"timeout" validate:"required,min=10,max=60000"`
	} `json:"http_client" validate:"required"`

	Logger struct {
		Level  string `json:"level" validate:"required,oneof=debug info warn error"`
		Format string `json:"format" validate:"required,oneof='2006-01-02 15:04:05.000'"`
	} `json:"logger" validate:"required"`
}

// New creates a new instance of Config and loads the default and local config files.
func New() (*Config, error) {
	c := &Config{}

	content, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return c, errors.WithStack(err)
	}
	err = json.Unmarshal(content, c)
	if err != nil {
		return c, errors.WithStack(err)
	}

	if utils.FileExist(localConfigPath) {
		content, err = os.ReadFile(localConfigPath)
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
		if err = os.WriteFile(localConfigPath, contentBytes, 0755); err != nil {
			return c, errors.WithStack(err)
		}
	}

	fmt.Println("config: loaded from file(s)", c.String())

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
