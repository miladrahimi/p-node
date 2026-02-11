package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/miladrahimi/p-node/pkg/util"
)

// Database is a simple file-backed JSON store.
type Database[T any] struct {
	data      *T
	directory string
	locker    *sync.Mutex
	validator *validator.Validate
}

// New creates a new database instance.
func New[T any](directory string, data *T) (*Database[T], error) {
	if directory == "" {
		return nil, errors.New("database: missing directory")
	}
	if data == nil {
		return nil, errors.New("database: missing data")
	}

	return &Database[T]{
		data:      data,
		directory: directory,
		locker:    &sync.Mutex{},
		validator: validator.New(validator.WithRequiredStructEnabled()),
	}, nil
}

// Init loads the data from the file if it exists, otherwise it creates the file with the default data.
func (d *Database[T]) Init() error {
	if util.FileExist(d.filePath()) {
		return d.Load()
	}

	return errors.WithStack(d.Save())
}

// Load loads the data from the file.
func (d *Database[T]) Load() error {
	d.locker.Lock()
	defer d.locker.Unlock()

	bytes, err := os.ReadFile(d.filePath())
	if err != nil {
		return errors.WithStack(err)
	}

	if err = json.Unmarshal(bytes, d.data); err != nil {
		return errors.WithStack(err)
	}

	return d.validator.Struct(d.data)
}

// Save saves the data to the file.
func (d *Database[T]) Save() error {
	d.locker.Lock()
	defer d.locker.Unlock()

	if err := os.MkdirAll(filepath.Dir(d.filePath()), 0o755); err != nil {
		return errors.WithStack(err)
	}

	bytes, err := json.Marshal(d.data)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.WriteFile(d.filePath(), bytes, 0644)
	return errors.WithStack(err)
}

// Backup writes a 7x24 rotating backup file based on the given directory template.
func (d *Database[T]) Backup() error {
	d.locker.Lock()
	bytes, err := json.Marshal(d.data)
	d.locker.Unlock()
	if err != nil {
		return errors.WithStack(err)
	}

	now := time.Now()
	path := strings.ToLower(fmt.Sprintf(d.backupPathTemplate(), now.Format("Mon"), now.Format("15")))
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return errors.WithStack(err)
	}

	if err = os.WriteFile(path, bytes, 0644); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Data returns the data.
func (d *Database[T]) Data() *T {
	return d.data
}

// path returns the path of the database file path.
func (d *Database[T]) filePath() string {
	return filepath.Join(d.directory, "data.json")
}

// backupPathTemplate returns the path of the backup file path template.
func (d *Database[T]) backupPathTemplate() string {
	return filepath.Join(d.directory, "backup-%s-%s.json")
}
