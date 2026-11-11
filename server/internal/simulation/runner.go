package simulation

import (
	"context"
	"fmt"
	"time"
)

// Runner advances a world at a fixed interval until its context is cancelled.
type Runner struct {
	world        *World
	interval     time.Duration
	tickObserver func(elapsed, budget time.Duration)
}

// SetTickObserver installs an inexpensive observer for operational metrics.
// Configure it before Run starts.
func (r *Runner) SetTickObserver(observer func(elapsed, budget time.Duration)) {
	r.tickObserver = observer
}

func NewRunner(world *World, interval time.Duration) (*Runner, error) {
	if world == nil {
		return nil, fmt.Errorf("world is required")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("tick interval must be positive")
	}

	return &Runner{world: world, interval: interval}, nil
}

func (r *Runner) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			startedAt := time.Now()
			r.world.Step(r.interval)
			if r.tickObserver != nil {
				r.tickObserver(time.Since(startedAt), r.interval)
			}
		}
	}
}
