package config

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/miladrahimi/xray-manager/pkg/utils"
	"os"
	"runtime"
)

const MainPath = "configs/main.json"
const LocalPath = "configs/main.local.json"
const AppName = "XrayNode"
const AppVersion = "v1.2.0"

var xrayConfigPath = "storage/xray.json"
var xrayBinaryPaths = map[string]string{
	"darwin": "third_party/xray-macos-arm64/xray",
	"linux":  "third_party/xray-linux-64/xray",
}

type Config struct {
	Logger struct {
		Level  string `json:"level"`
		Format string `json:"format"`
	} `json:"logger"`
}

func (c *Config) Init() (err error) {
	var content []byte
	if utils.FileExist(LocalPath) {
		content, err = os.ReadFile(LocalPath)
	} else {
		content, err = os.ReadFile(MainPath)
	}
	if err != nil {
		return fmt.Errorf("config: cannot load file, err: %v", err)
	}

	err = json.Unmarshal(content, &c)
	if err != nil {
		return fmt.Errorf("config: cannot unmarshal file, err: %v", err)
	}

	if err = validator.New().Struct(c); err != nil {
		return fmt.Errorf("config: cannot validate data, err: %v", err)
	}

	return nil
}

func (c *Config) XrayBinaryPath() string {
	if path, found := xrayBinaryPaths[runtime.GOOS]; found {
		return path
	}
	return xrayBinaryPaths["linux"]
}

func (c *Config) XrayConfigPath() string {
	return xrayConfigPath
}

func New() *Config {
	return &Config{}
}
