package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/miladrahimi/p-node/pkg/logger"
)

// Worker represents a worker that runs a function at a specified interval.
type Worker struct {
	name     string
	interval time.Duration
	body     func()
	logger   *logger.Logger
}

// New creates a new worker.
func New(name string, l *logger.Logger, interval time.Duration, body func()) *Worker {
	return &Worker{name: name, logger: l, interval: interval, body: body}
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
				w.body()
			}
		}
	}()
}
