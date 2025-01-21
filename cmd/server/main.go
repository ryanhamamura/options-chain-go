package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/joho/godotenv"
    "github.com/gorilla/mux"
    "github.com/ryanhamamura/options-chain-go/internal/api"
    "github.com/ryanhamamura/options-chain-go/internal/models"
    "github.com/ryanhamamura/options-chain-go/internal/stream"
    "github.com/ryanhamamura/options-chain-go/internal/tasty"
)

func main() {
    // Load .env file 
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }

    // Debug: Print environment variables (without password) 
    log.Printf("Environment: %s", os.Getenv("TASTY_ENVIRONMENT"))
    log.Printf("Username: %s", os.Getenv("TASTY_USERNAME"))
    log.Printf("Base URL: %s", os.Getenv("TASTY_BASE_URL"))

    // Load configuration
    config, err := tasty.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Log environment
    log.Printf("Starting in %s environment", config.Environment)
    if config.Environment == tasty.Sandbox {
        log.Println("Using sandbox URLs:")
        log.Printf("  API URL: %s", config.BaseURL)
        log.Printf("  Streamer URL: %s", config.StreamerURL)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create Tastytrade client with configuration
    client := tasty.NewClient(*config)
    
    // Configure error handlers
    client.SetErrorHandler(func(err error) {
        log.Printf("DXLink error: %v", err)
    })
    client.SetDisconnectHandler(func() {
        log.Println("DXLink disconnected, attempting reconnection...")
    })
    client.SetReconnectHandler(func() {
        log.Println("DXLink successfully reconnected")
    })

    // Get authentication credentials
    username := os.Getenv("TASTY_USERNAME")
    password := os.Getenv("TASTY_PASSWORD")
    sessionToken := os.Getenv("TASTY_SESSION_TOKEN")

    // If no session token is provided, login to get one
    if sessionToken == "" {
        if username == "" || password == "" {
            log.Fatal("Either TASTY_SESSION_TOKEN or both TASTY_USERNAME and TASTY_PASSWORD must be provided")
        }

        var err error
        sessionToken, err = client.Login(ctx, username, password)
        if err != nil {
            log.Fatalf("Failed to login: %v", err)
        }
        log.Println("Successfully logged in")
    } else {
        log.Println("Using provided session token")
    } 

    // Set the session token in the client
    client.SetSessionToken(sessionToken)

    // Get quote token
    _, err = client.GetQuoteToken(ctx)
    if err != nil {
        log.Fatalf("Failed to get quote token: %v", err)
    }
    log.Println("Successfully obtained quote token")

    // Connect to DXLink
    if err := client.ConnectDXLink(ctx); err != nil {
        log.Fatalf("Failed to connect to DXLink: %v", err)
    }
    defer client.Close()

    // Create WebSocket manager for our frontend
    wsManager := stream.NewManager()

    // Create router and handler
    r := mux.NewRouter()
    handler := api.NewHandler(wsManager)
    api.SetupRoutes(r, handler)

    // Subscribe to market data for specific options
    subscriptions := []tasty.DXSubscription{
        {Type: "Quote", Symbol: "SPY"},
        {Type: "Greeks", Symbol: "SPY"},
        {Type: "Trade", Symbol: "SPY"},
        {Type: "Summary", Symbol: "SPY"},
    }
    if err := client.Subscribe(ctx, 1, subscriptions); err != nil {
        log.Fatalf("Failed to subscribe: %v", err)
    }

    // Start reading market data with automatic reconnection
    client.StartReading(ctx, func(chain models.OptionChain) {
        // Broadcast the transformed data to all connected frontend clients
        wsManager.BroadcastOptionChain(chain)
    })

    // Create server with timeouts from config
    port := getEnvOrDefault("APP_PORT", "8080")
    server := &http.Server{
        Addr:         ":" + port,
        Handler:      r,
        WriteTimeout: config.WSWriteTimeout,
        ReadTimeout:  config.WSReadTimeout,
        IdleTimeout:  time.Minute,
    }

    // Handle graceful shutdown
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
    
    // Start server in goroutine
    go func() {
        log.Printf("Server starting on port %s", port)
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Wait for shutdown signal
    <-shutdown
    log.Println("Shutting down gracefully...")
    
    // Create shutdown context with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownCancel()

    // Shutdown server
    if err := server.Shutdown(shutdownCtx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }
    
    // Cancel main context to stop client operations
    cancel()

    log.Println("Server stopped")
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}



