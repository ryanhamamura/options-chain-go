package main

import (
    "log"
    "math/rand"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/ryanhamamura/options-chain-go/internal/api"
    "github.com/ryanhamamura/options-chain-go/internal/stream"
)

func main() {
    // Initialize random seed 
    rand.Seed(time.Now().UnixNano())

    // Create WebSocket manager
    wsManager := stream.NewManager()

    // Create router
    r := mux.NewRouter()

    // Create handler
    handler := api.NewHandler(wsManager)

    // Setup routes
    api.SetupRoutes(r, handler)

    // Start the streaming simulation in the background
    go wsManager.StartSimulation()

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
