package database

import (
	"encoding/json"
	"github.com/go-playground/validator"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/xray-manager/pkg/logger"
	"github.com/miladrahimi/xray-manager/pkg/utils"
	"go.uber.org/zap"
	"math/rand"
	"os"
)

const Path = "storage/database/app.json"

type Data struct {
	Settings *Settings `json:"settings"`
}

type Database struct {
	Data *Data
	l    *logger.Logger
}

func (d *Database) Init() {
	if !utils.FileExist(Path) {
		if !utils.PortFree(d.Data.Settings.HttpPort) {
			var err error
			if d.Data.Settings.HttpPort, err = utils.FreePort(); err != nil {
				d.l.Exit("database: cannot init free port for http", zap.Error(err))
			}
		}
		d.Save()
	}
	d.Load()
}

func (d *Database) Load() {
	content, err := os.ReadFile(Path)
	if err != nil {
		d.l.Exit("database: cannot load file", zap.String("file", Path), zap.Error(err))
	}

	err = json.Unmarshal(content, d.Data)
	if err != nil {
		d.l.Exit("database: cannot unmarshall data", zap.Error(err))
	}

	if err = validator.New().Struct(d); err != nil {
		d.l.Exit("database: cannot validate data", zap.Error(err))
	}
}

func (d *Database) Save() {
	defer func() {
		d.Load()
	}()
	content, err := json.Marshal(d.Data)
	if err != nil {
		d.l.Exit("database: cannot marshal data", zap.Error(err))
	}

	if err = os.WriteFile(Path, content, 0755); err != nil {
		d.l.Exit("database: cannot save file", zap.String("file", Path), zap.Error(err))
	}
}

func New(l *logger.Logger) *Database {
	return &Database{
		l: l,
		Data: &Data{
			Settings: &Settings{
				HttpPort:  rand.Intn(64536) + 1000,
				HttpToken: random.String(16),
			},
		},
	}
}
