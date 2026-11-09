package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterDeniesUntilWindowExpires(t *testing.T) {
	now := time.Date(2026, time.November, 9, 12, 0, 0, 0, time.UTC)
	limiter, err := New(Config{
		Limit: 2, Window: 10 * time.Second, MaxEntries: 10,
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if allowed, _ := limiter.Allow("player"); !allowed {
			t.Fatalf("attempt %d was denied", attempt+1)
		}
	}
	if allowed, retryAfter := limiter.Allow("player"); allowed || retryAfter != 10*time.Second {
		t.Fatalf("third attempt = allowed %t, retry after %s", allowed, retryAfter)
	}
	now = now.Add(10 * time.Second)
	if allowed, _ := limiter.Allow("player"); !allowed {
		t.Fatal("attempt after the window was denied")
	}
}

func TestLimiterResetForgivesFailures(t *testing.T) {
	limiter, err := New(Config{Limit: 1, Window: time.Minute, MaxEntries: 10})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if allowed, _ := limiter.Allow("player"); !allowed {
		t.Fatal("first attempt was denied")
	}
	limiter.Reset("player")
	if allowed, _ := limiter.Allow("player"); !allowed {
		t.Fatal("attempt after reset was denied")
	}
}

func TestLimiterBoundsTrackedClients(t *testing.T) {
	now := time.Now()
	limiter, err := New(Config{
		Limit: 1, Window: time.Hour, MaxEntries: 2,
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	limiter.Allow("one")
	now = now.Add(time.Second)
	limiter.Allow("two")
	now = now.Add(time.Second)
	limiter.Allow("three")
	if len(limiter.entries) != 2 {
		t.Fatalf("tracked entries = %d, want 2", len(limiter.entries))
	}
	if _, exists := limiter.entries["one"]; exists {
		t.Fatal("oldest entry was not evicted")
	}
}
