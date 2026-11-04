package simulation

import (
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
