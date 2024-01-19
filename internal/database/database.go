package database

import (
	"encoding/json"
	"github.com/go-playground/validator"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/xray-manager/pkg/utils"
	"go.uber.org/zap"
	"os"
)

const Path = "storage/database.json"

type Data struct {
	Settings *Settings `json:"settings"`
}

type Database struct {
	Data *Data
	log  *zap.Logger
}

func (d *Database) Init() {
	if !utils.FileExist(Path) {
		d.Save()
	}
	d.Load()
}

func (d *Database) Load() {
	content, err := os.ReadFile(Path)
	if err != nil {
		d.log.Fatal("database: cannot load file", zap.String("file", Path), zap.Error(err))
	}

	err = json.Unmarshal(content, d.Data)
	if err != nil {
		d.log.Fatal("database: cannot unmarshall data", zap.Error(err))
	}

	if err = validator.New().Struct(d); err != nil {
		d.log.Fatal("database: cannot validate data", zap.Error(err))
	}
}

func (d *Database) Save() {
	defer func() {
		d.Load()
	}()
	content, err := json.Marshal(d.Data)
	if err != nil {
		d.log.Fatal("database: cannot marshal data", zap.Error(err))
	}

	if err = os.WriteFile(Path, content, 0755); err != nil {
		d.log.Fatal("database: cannot save file", zap.String("file", Path), zap.Error(err))
	}
}

func New(l *zap.Logger) *Database {
	return &Database{
		log: l,
		Data: &Data{
			Settings: &Settings{
				XrayApiPort: 3411,
				HttpPort:    1826,
				HttpToken:   random.String(16),
			},
		},
	}
}
