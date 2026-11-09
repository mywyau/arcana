// Package ratelimit provides a small bounded fixed-window attempt limiter.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

type Config struct {
	Limit      int
	Window     time.Duration
	MaxEntries int
	Now        func() time.Time
}

type entry struct {
	count       int
	windowStart time.Time
	lastSeen    time.Time
}

type Limiter struct {
	mu         sync.Mutex
	limit      int
	window     time.Duration
	maxEntries int
	now        func() time.Time
	entries    map[string]entry
}

func New(config Config) (*Limiter, error) {
	if config.Limit < 1 {
		return nil, fmt.Errorf("rate limit must be positive")
	}
	if config.Window <= 0 {
		return nil, fmt.Errorf("rate limit window must be positive")
	}
	if config.MaxEntries < 1 {
		return nil, fmt.Errorf("maximum rate limit entries must be positive")
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	return &Limiter{
		limit: config.Limit, window: config.Window, maxEntries: config.MaxEntries,
		now: now, entries: make(map[string]entry),
	}, nil
}

// Allow consumes one attempt. A denied attempt does not extend the window.
func (limiter *Limiter) Allow(key string) (bool, time.Duration) {
	if key == "" {
		key = "unknown"
	}
	now := limiter.now()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	value, found := limiter.entries[key]
	if found && now.Sub(value.windowStart) >= limiter.window {
		delete(limiter.entries, key)
		found = false
	}
	if found && value.count >= limiter.limit {
		return false, max(value.windowStart.Add(limiter.window).Sub(now), time.Second)
	}
	if !found {
		limiter.makeRoom(now)
		value = entry{windowStart: now}
	}
	value.count++
	value.lastSeen = now
	limiter.entries[key] = value
	return true, 0
}

// Blocked reports whether a key has exhausted its attempts without consuming
// another attempt. Expired windows are removed.
func (limiter *Limiter) Blocked(key string) (bool, time.Duration) {
	if key == "" {
		key = "unknown"
	}
	now := limiter.now()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	value, found := limiter.entries[key]
	if !found {
		return false, 0
	}
	if now.Sub(value.windowStart) >= limiter.window {
		delete(limiter.entries, key)
		return false, 0
	}
	if value.count < limiter.limit {
		return false, 0
	}
	return true, max(value.windowStart.Add(limiter.window).Sub(now), time.Second)
}

// Reset forgives attempts after a successful operation.
func (limiter *Limiter) Reset(key string) {
	if key == "" {
		key = "unknown"
	}
	limiter.mu.Lock()
	delete(limiter.entries, key)
	limiter.mu.Unlock()
}

func (limiter *Limiter) makeRoom(now time.Time) {
	if len(limiter.entries) < limiter.maxEntries {
		return
	}
	for key, value := range limiter.entries {
		if now.Sub(value.lastSeen) >= limiter.window {
			delete(limiter.entries, key)
		}
	}
	if len(limiter.entries) < limiter.maxEntries {
		return
	}
	var oldestKey string
	var oldestSeen time.Time
	for key, value := range limiter.entries {
		if oldestKey == "" || value.lastSeen.Before(oldestSeen) {
			oldestKey = key
			oldestSeen = value.lastSeen
		}
	}
	delete(limiter.entries, oldestKey)
}
