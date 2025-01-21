package tasty

import (
    "sync"
    "time"
)

// ReconnectConfig holds configuration for reconnection attempts
type ReconnectConfig struct {
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    MaxAttempts   int
    ResetAfter    time.Duration
}

// DefaultReconnectConfig provides sensible defaults
var DefaultReconnectConfig = ReconnectConfig{
    InitialDelay: time.Second,
    MaxDelay:     time.Minute,
    MaxAttempts:  10,
    ResetAfter:   time.Minute * 5,
}

type reconnectManager struct {
    config ReconnectConfig
    mu     sync.Mutex

    attempts     int
    lastAttempt  time.Time
    currentDelay time.Duration
}

func newReconnectManager(config ReconnectConfig) *reconnectManager {
    return &reconnectManager{
        config:       config,
        currentDelay: config.InitialDelay,
    }
}

func (r *reconnectManager) nextDelay() (time.Duration, bool) {
    r.mu.Lock()
    defer r.mu.Unlock()

    now := time.Now()

    // Reset attempts if enough time has passed
    if now.Sub(r.lastAttempt) > r.config.ResetAfter {
        r.attempts = 0
        r.currentDelay = r.config.InitialDelay
    }

    // Check if we've exceeded max attempts
    if r.attempts >= r.config.MaxAttempts {
        return 0, false
    }

    r.attempts++
    r.lastAttempt = now

    // Exponential backoff
    if r.attempts > 1 {
        r.currentDelay *= 2
        if r.currentDelay > r.config.MaxDelay {
            r.currentDelay = r.config.MaxDelay
        }
    }

    return r.currentDelay, true
}

func (r *reconnectManager) reset() {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.attempts = 0
    r.currentDelay = r.config.InitialDelay
    r.lastAttempt = time.Time{}
}
