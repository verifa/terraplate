package runner

import (
	"context"
	"errors"
)

func (r *Runner) startWorker(workerID int) {
	for run := range r.runQueue {
		// Check if context has been cancelled
		if errors.Is(r.ctx.Err(), context.Canceled) {
			run.Cancelled = true
		} else {
			run.Run()
		}
		r.wg.Done()
	}
}
