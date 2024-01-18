package database

import (
	"encoding/json"
	"github.com/go-playground/validator"
	"github.com/labstack/gommon/random"
	"github.com/miladrahimi/xray-manager/pkg/utils"
	"go.uber.org/zap"
	"os"
	"sync"
)

const Path = "storage/database.json"

type Data struct {
	Settings *Settings `json:"settings"`
}

type Database struct {
	Data   *Data
	Locker sync.Mutex
	log    *zap.Logger
}

func (d *Database) Init() {
	d.Locker.Lock()
	defer d.Locker.Unlock()

	if !utils.FileExist(Path) {
		d.initData()
		d.Save()
	}
	d.Load()
}

func (d *Database) initData() {
	var err error
	if d.Data.Settings.InternalPort, err = utils.FreePort(); err != nil {
		d.log.Fatal("database: cannot init Settings.InternalPort", zap.Error(err))
	}
	if d.Data.Settings.HttpPort, err = utils.FreePort(); err != nil {
		d.log.Fatal("database: cannot init Settings.HttpPort", zap.Error(err))
	}
	d.Data.Settings.HttpToken = random.String(16)
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
			Settings: &Settings{},
		},
	}
}
