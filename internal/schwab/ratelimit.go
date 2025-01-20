package schwab

import (
    "context"
    "sync"
    "time"
)

// RateLimiter implements rate limiting for API requests
type RateLimiter struct {
    mutex      sync.Mutex
    lastAccess time.Time
    interval   time.Duration
}

// NewRateLimiter creates a new rate limiter with specified interval
func NewRateLimiter(interval time.Duration) *RateLimiter {
    return &RateLimiter{
        interval: interval,
    }
}

// Wait blocks until rate limit allows a new request
func (r *RateLimiter) Wait(ctx context.Context) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    if !r.lastAccess.IsZero() {
        waitTime := r.interval - time.Since(r.lastAccess)
        if waitTime > 0 {
            select {
            case <-time.After(waitTime):
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }

    r.lastAccess = time.Now()
    return nil
}
