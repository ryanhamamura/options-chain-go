package tasty

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

// Environment represents the Tastytrade environment
type Environment string

const (
    Sandbox     Environment = "sandbox"
    Production  Environment = "production"
)

// URLs for different environments
const (
    SandboxBaseURL     = "https://api.cert.tastyworks.com"
    SandboxStreamerURL = "wss://streamer.cert.tastyworks.com"
    
    ProductionBaseURL     = "https://api.tastytrade.com"
    ProductionStreamerURL = "wss://streamer.tastytrade.com"
)

// Config holds the client configuration
type Config struct {
    // Environment
    Environment Environment
    BaseURL     string
    StreamerURL string
    
    // Authentication
    SessionToken string

    // WebSocket settings
    WSPingInterval time.Duration
    WSPingTimeout  time.Duration
    WSWriteTimeout time.Duration
    WSReadTimeout  time.Duration

    // Rate limiting
    RateLimitRequests int
    RateLimitInterval time.Duration

    // Reconnection settings
    ReconnectInitialDelay time.Duration
    ReconnectMaxDelay     time.Duration
    ReconnectMaxAttempts  int
    ReconnectResetAfter   time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
    config := &Config{}

    // Load environment
    env := Environment(getEnvOrDefault("TASTY_ENVIRONMENT", string(Sandbox)))
    if env != Sandbox && env != Production {
        return nil, fmt.Errorf("invalid environment: %s", env)
    }
    config.Environment = env

    // Set URLs based on environment
    switch env {
    case Sandbox:
        config.BaseURL = getEnvOrDefault("TASTY_BASE_URL", SandboxBaseURL)
        config.StreamerURL = getEnvOrDefault("TASTY_STREAMER_URL", SandboxStreamerURL)
    case Production:
        config.BaseURL = getEnvOrDefault("TASTY_BASE_URL", ProductionBaseURL)
        config.StreamerURL = getEnvOrDefault("TASTY_STREAMER_URL", ProductionStreamerURL)
    }

    // Load authentication
    config.SessionToken = os.Getenv("TASTY_SESSION_TOKEN")
    if config.SessionToken == "" {
        // Check if username/password are provided instead
        username := os.Getenv("TASTY_USERNAME")
        password := os.Getenv("TASTY_PASSWORD")
        if username == "" || password == "" {
            return nil, fmt.Errorf("Either TASTY_SESSION_TOKEN or both TASTY_USERNAME and TASTY_PASSWORD must be provided")
        }
    }

    // Load WebSocket settings
    config.WSPingInterval = getDurationOrDefault("WS_PING_INTERVAL", 30*time.Second)
    config.WSPingTimeout = getDurationOrDefault("WS_PING_TIMEOUT", 10*time.Second)
    config.WSWriteTimeout = getDurationOrDefault("WS_WRITE_TIMEOUT", 15*time.Second)
    config.WSReadTimeout = getDurationOrDefault("WS_READ_TIMEOUT", 15*time.Second)

    // Load rate limiting settings
    config.RateLimitRequests = getIntOrDefault("RATE_LIMIT_REQUESTS", 10)
    config.RateLimitInterval = getDurationOrDefault("RATE_LIMIT_INTERVAL", time.Second)

    // Load reconnection settings
    config.ReconnectInitialDelay = getDurationOrDefault("RECONNECT_INITIAL_DELAY", time.Second)
    config.ReconnectMaxDelay = getDurationOrDefault("RECONNECT_MAX_DELAY", time.Minute)
    config.ReconnectMaxAttempts = getIntOrDefault("RECONNECT_MAX_ATTEMPTS", 10)
    config.ReconnectResetAfter = getDurationOrDefault("RECONNECT_RESET_AFTER", 5*time.Minute)

    return config, validateConfig(config)
}

// Helper functions to get environment variables with defaults
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
    str := os.Getenv(key)
    if str == "" {
        return defaultValue
    }
    val, err := strconv.Atoi(str)
    if err != nil {
        return defaultValue
    }
    return val
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
    str := os.Getenv(key)
    if str == "" {
        return defaultValue
    }
    duration, err := time.ParseDuration(str)
    if err != nil {
        return defaultValue
    }
    return duration
}

func validateConfig(config *Config) error {
    if config.BaseURL == "" {
        return fmt.Errorf("base URL is required")
    }
    if config.StreamerURL == "" {
        return fmt.Errorf("streamer URL is required")
    }
    if config.WSPingInterval <= 0 {
        return fmt.Errorf("invalid websocket ping interval")
    }
    if config.RateLimitRequests <= 0 {
        return fmt.Errorf("invalid rate limit requests")
    }
    if config.ReconnectMaxAttempts <= 0 {
        return fmt.Errorf("invalid max reconnection attempts")
    }
    return nil
}
