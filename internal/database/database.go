package database

import (
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/p-node/internal/utils"
	"github.com/miladrahimi/p-node/pkg/logger"
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
	locker *sync.Mutex
	Data   *Data
}

func (d *Database) Init() {
	d.locker.Lock()
	defer d.locker.Unlock()

	if !utils.FileExist(Path) {
		if !utils.PortFree(d.Data.Settings.HttpPort) {
			var err error
			if d.Data.Settings.HttpPort, err = utils.FreePort(); err != nil {
				d.l.Fatal("database: cannot find port for http", zap.Error(errors.WithStack(err)))
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
		d.l.Fatal("database: cannot load file", zap.String("file", Path), zap.Error(errors.WithStack(err)))
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
		locker: &sync.Mutex{},
		l:      l,
		Data: &Data{
			Settings: &Settings{
				HttpPort:  rand.Intn(64536) + 1000,
				HttpToken: random.String(16),
			},
		},
	}
}
