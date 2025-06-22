package worker

import (
	"context"
	"time"
)

type Worker struct {
	context  context.Context
	interval time.Duration
	body     func()
	stop     func()
}

func (w *Worker) Start() {
	ticker := time.NewTicker(w.interval)
	go func() {
		for {
			select {
			case <-w.context.Done():
				w.stop()
				return
			case <-ticker.C:
				w.body()
			}
		}
	}()
}

func New(c context.Context, interval time.Duration, body func(), stop func()) *Worker {
	return &Worker{context: c, interval: interval, body: body, stop: stop}
}
