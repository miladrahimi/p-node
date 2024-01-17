package config

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"os"
	"xray-node/internal/utils"
)

const MainPath = "configs/main.json"
const LocalPath = "configs/main.local.json"
const AppName = "XrayNode"
const AppVersion = "v1.0.0"
const ShadowsocksMethod = "chacha20-ietf-poly1305"

// Config is the root configuration.
type Config struct {
	Logger struct {
		Level  string `json:"level"`
		Format string `json:"format"`
	} `json:"logger"`

	Worker struct {
		Interval int `json:"interval" validate:"required,min=10"`
	} `json:"worker"`
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

// New creates an instance of the Config.
func New() *Config {
	return &Config{}
}
