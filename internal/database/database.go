package database

import (
	"encoding/json"
	"github.com/go-playground/validator"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/p-manager/pkg/logger"
	"github.com/miladrahimi/p-manager/pkg/utils"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"sync"
)

const Path = "storage/database/app.json"

type Data struct {
	Settings *Settings `json:"settings"`
}

type Database struct {
	l      *logger.Logger
	Locker *sync.Mutex
	Data   *Data
}

func (d *Database) Init() {
	d.Locker.Lock()
	defer d.Locker.Unlock()

	if !utils.FileExist(Path) {
		if !utils.PortFree(d.Data.Settings.HttpPort) {
			var err error
			if d.Data.Settings.HttpPort, err = utils.FreePort(); err != nil {
				d.l.Fatal("database: cannot init free port for http", zap.Error(err))
			}
		}
		d.Save()
	} else {
		d.Load()
	}
}

func (d *Database) Load() {
	content, err := os.ReadFile(Path)
	if err != nil {
		d.l.Fatal("database: cannot load file", zap.String("file", Path), zap.Error(err))
	}

	err = json.Unmarshal(content, d.Data)
	if err != nil {
		d.l.Fatal("database: cannot unmarshall data", zap.Error(err))
	}

	if err = validator.New().Struct(d); err != nil {
		d.l.Fatal("database: cannot validate data", zap.Error(err))
	}
}

func (d *Database) Save() {
	content, err := json.Marshal(d.Data)
	if err != nil {
		d.l.Fatal("database: cannot marshal data", zap.Error(err))
	}

	if err = os.WriteFile(Path, content, 0755); err != nil {
		d.l.Fatal("database: cannot save file", zap.String("file", Path), zap.Error(err))
	}
}

func New(l *logger.Logger) *Database {
	return &Database{
		Locker: &sync.Mutex{},
		l:      l,
		Data: &Data{
			Settings: &Settings{
				HttpPort:  rand.Intn(64536) + 1000,
				HttpToken: random.String(16),
			},
		},
	}
}
