package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/miladrahimi/p-node/pkg/logger"
	"go.uber.org/zap"
)

type Task func() error

// Worker represents a worker that runs a function at a specified interval.
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

// Start starts the worker.
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				w.logger.Info(fmt.Sprintf("coordinator: worker '%s': stopped", w.name))
				return
			case <-ticker.C:
				w.logger.Info(fmt.Sprintf("coordinator: worker '%s': running...", w.name))
				if err := w.task(); err != nil {
					w.logger.Error(
						fmt.Sprintf("coordinator: worker '%s': error", w.name),
						zap.Error(err),
					)
				}
			}
		}
	}()
}
