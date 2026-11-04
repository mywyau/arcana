package simulation

import (
	"context"
	"fmt"
	"time"
)

// Runner advances a world at a fixed interval until its context is cancelled.
type Runner struct {
	world    *World
	interval time.Duration
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
			r.world.Step(r.interval)
		}
	}
}
