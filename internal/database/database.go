package database

import (
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/p-node/internal/utils"
	"github.com/miladrahimi/p-node/pkg/logger"
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

func (d *Database) Init() error {
	d.locker.Lock()
	defer d.locker.Unlock()

	if utils.FileExist(Path) {
		return d.Load()
	}

	if !utils.PortFree(d.Data.Settings.HttpPort) {
		var err error
		if d.Data.Settings.HttpPort, err = utils.FreePort(); err != nil {
			return errors.Wrap(err, "cannot find free port")
		}
	}

	err := d.Save()
	return errors.WithStack(err)
}

func (d *Database) Load() error {
	content, err := os.ReadFile(Path)
	if err != nil {
		return errors.WithStack(err)
	}

	err = json.Unmarshal(content, d.Data)
	if err != nil {
		return errors.WithStack(err)
	}

	err = validator.New().Struct(d)
	return errors.WithStack(err)
}

func (d *Database) Save() error {
	content, err := json.Marshal(d.Data)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.WriteFile(Path, content, 0755)
	return errors.WithStack(err)
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
