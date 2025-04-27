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
const envConfigPath = "configs/main.json"

const AppName = "P-Node"
const AppVersion = "v1.4.0"

const XrayConfigPath = "storage/app/xray.json"
const XrayLogLevel = "debug"

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
		Level  string `json:"level"`
		Format string `json:"format"`
	} `json:"logger"`
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

	if utils.FileExist(envConfigPath) {
		content, err = os.ReadFile(envConfigPath)
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
