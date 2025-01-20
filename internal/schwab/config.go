package schwab

import (
    "fmt"
    "os"
)

// Config holds configuration for the Schwab client
type Config struct {
    BaseURL    string
    WSURL      string
    APIKey     string
    APISecret  string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
    config := &Config{
        BaseURL:    getEnvOrDefault("SCHWAB_API_URL", "https://api.schwab.com"),
        WSURL:      getEnvOrDefault("SCHWAB_WS_URL", "wss://stream.schwab.com"),
        APIKey:     os.Getenv("SCHWAB_API_KEY"),
        APISecret:  os.Getenv("SCHWAB_API_SECRET"),
    }

    if config.APIKey == "" || config.APISecret == "" {
        return nil, fmt.Errorf("missing required environment variables: SCHWAB_API_KEY and/or SCHWAB_API_SECRET")
    }

    return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
