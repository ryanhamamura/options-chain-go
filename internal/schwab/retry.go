package schwab

import (
    "context"
    "math"
    "time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
    MaxRetries  int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

// DefaultRetryConfig provides default retry settings
var DefaultRetryConfig = RetryConfig{
    MaxRetries: 3,
    BaseDelay:  100 * time.Millisecond,
    MaxDelay:   2 * time.Second,
}

// retry executes the given function with exponential backoff
func retry(ctx context.Context, config RetryConfig, fn func() error) error {
    var err error
    
    for attempt := 0; attempt <= config.MaxRetries; attempt++ {
        if err = fn(); err == nil {
            return nil
        }

        if attempt == config.MaxRetries {
            break
        }

        delay := time.Duration(math.Min(
            float64(config.BaseDelay)*math.Pow(2, float64(attempt)),
            float64(config.MaxDelay),
        ))

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
        }
    }

    return err
}
