package database

import (
	"encoding/json"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/p-node/internal/utils"
	"github.com/miladrahimi/p-node/pkg/logger"
)

const Path = "storage/database/app.json"

// Schema is the database schema.
type Schema struct {
	Settings *Settings `json:"settings"`
	Manager  *Manager  `json:"manager"`
}

// Database is the database struct.
type Database struct {
	l      *logger.Logger
	locker *sync.Mutex
	Data   *Schema
}

// New creates a new instance of Database.
func New(l *logger.Logger) *Database {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Database{
		locker: &sync.Mutex{},
		l:      l,
		Data: &Schema{
			Manager: nil,
			Settings: &Settings{
				HttpPort:  r.Intn(64536) + 1000,
				HttpToken: random.String(16),
			},
		},
	}
}

// Init initializes the database.
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

// Load loads the database from the file.
func (d *Database) Load() error {
	content, err := os.ReadFile(Path)
	if err != nil {
		return errors.WithStack(err)
	}

	err = json.Unmarshal(content, d.Data)
	if err != nil {
		return errors.WithStack(err)
	}

	v := validator.New()
	if d.Data == nil {
		return errors.New("database: missing schema")
	}
	if d.Data.Settings == nil {
		return errors.New("database: missing settings")
	}
	if err = v.Struct(d.Data.Settings); err != nil {
		return errors.WithStack(err)
	}
	if d.Data.Manager != nil {
		if err = v.Struct(d.Data.Manager); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// Save saves the database to the file.
func (d *Database) Save() error {
	content, err := json.Marshal(d.Data)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.WriteFile(Path, content, 0644)
	return errors.WithStack(err)
}
