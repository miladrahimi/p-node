package config

import (
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/miladrahimi/p-node/internal/utils"
	"os"
	"runtime"
)

const defaultConfigPath = "configs/main.defaults.json"
const configPath = "configs/main.json"

const AppName = "P-Node"
const AppVersion = "v25.6.22"

const XrayConfigPath = "storage/app/xray.json"
const XrayLogLevel = "debug"

const HttpTimeout = 20

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
	Logger struct {
		Level  string `json:"level" validate:"required,oneof=debug info warn error"`
		Format string `json:"format" validate:"required,oneof='2006-01-02 15:04:05.000'"`
	} `json:"logger" validate:"required"`
}

func (c *Config) toString() (string, error) {
	j, err := json.Marshal(c)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(j), nil
}

func (c *Config) Init() (err error) {
	content, err := os.ReadFile(defaultConfigPath)
	if err != nil {
		return errors.WithStack(err)
	}
	err = json.Unmarshal(content, &c)
	if err != nil {
		return errors.WithStack(err)
	}

	if utils.FileExist(configPath) {
		content, err = os.ReadFile(configPath)
		if err != nil {
			return errors.WithStack(err)
		}
		if err = json.Unmarshal(content, &c); err != nil {
			return errors.WithStack(err)
		}
	}

	configString, err := c.toString()
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Println("Config:", configString)

	return errors.WithStack(validator.New().Struct(c))
}

func New() *Config {
	return &Config{}
}
