package simulation

import (
	"context"
	"testing"
	"time"
)

func TestNewRunnerRejectsInvalidConfiguration(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name     string
		world    *World
		interval time.Duration
	}{
		{name: "missing world", interval: time.Second},
		{name: "zero interval", world: world},
		{name: "negative interval", world: world, interval: -time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewRunner(tt.world, tt.interval); err == nil {
				t.Fatal("NewRunner() error = nil, want an error")
			}
		})
	}
}

func TestNewRunnerAcceptsValidConfiguration(t *testing.T) {
	if _, err := NewRunner(NewWorld(), time.Second); err != nil {
		t.Fatalf("NewRunner() returned an unexpected error: %v", err)
	}
}

func TestRunnerReportsTickDurationAndBudget(t *testing.T) {
	const interval = time.Millisecond
	runner, err := NewRunner(NewWorld(), interval)
	if err != nil {
		t.Fatalf("NewRunner() returned an unexpected error: %v", err)
	}
	reported := make(chan struct {
		elapsed time.Duration
		budget  time.Duration
	}, 1)
	runner.SetTickObserver(func(elapsed, budget time.Duration) {
		select {
		case reported <- struct {
			elapsed time.Duration
			budget  time.Duration
		}{elapsed: elapsed, budget: budget}:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runner.Run(ctx)

	select {
	case sample := <-reported:
		if sample.elapsed < 0 {
			t.Fatalf("reported elapsed duration = %s", sample.elapsed)
		}
		if sample.budget != interval {
			t.Fatalf("reported budget = %s, want %s", sample.budget, interval)
		}
	case <-time.After(time.Second):
		t.Fatal("runner did not report a tick")
	}
}
