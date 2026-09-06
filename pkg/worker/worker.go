package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/miladrahimi/p-node/pkg/logger"
	"go.uber.org/zap"
)

// Task is the unit of work a worker runs on every tick.
type Task func() error

// Worker runs a task at a fixed interval until its context is canceled.
type Worker struct {
	name     string
	interval time.Duration
	task     Task
	logger   *logger.Logger
}

// New creates a new worker.
func New(name string, l *logger.Logger, interval time.Duration, task Task) *Worker {
	return &Worker{name: name, logger: l, interval: interval, task: task}
}

// Start starts the worker in the background.
func (w *Worker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				w.logger.Info(fmt.Sprintf("worker '%s': stopped", w.name))
				return
			case <-ticker.C:
				w.logger.Debug(fmt.Sprintf("worker '%s': running...", w.name))
				w.run()
			}
		}
	}()
}

// run executes the task once, logging errors and recovering from panics so one
// bad tick does not take the process down.
func (w *Worker) run() {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error(fmt.Sprintf("worker '%s': recovered from panic: %v", w.name, r))
		}
	}()
	if err := w.task(); err != nil {
		w.logger.Error(fmt.Sprintf("worker '%s': error", w.name), zap.Error(err))
	}
}
