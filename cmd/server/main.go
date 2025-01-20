// main.go
package main

import (
    "encoding/json"
    "log"
    "math/rand"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

// OptionData represents a single option contract
type OptionData struct {
    Strike      float64 `json:"strike"`
    Expiration  string  `json:"expiration"`
    Type        string  `json:"type"` // "call" or "put"
    Bid         float64 `json:"bid"`
    Ask         float64 `json:"ask"`
    LastPrice   float64 `json:"lastPrice"`
    Volume      int     `json:"volume"`
    OpenInt     int     `json:"openInterest"`
    Delta       float64 `json:"delta"`
    Gamma       float64 `json:"gamma"`
    Theta       float64 `json:"theta"`
    Vega        float64 `json:"vega"`
    ImpliedVol  float64 `json:"impliedVolatility"`
}

// OptionChain represents the full options chain
type OptionChain struct {
    Symbol      string       `json:"symbol"`
    Underlying  float64      `json:"underlyingPrice"`
    Updated     time.Time    `json:"lastUpdated"`
    Calls       []OptionData `json:"calls"`
    Puts        []OptionData `json:"puts"`
}

var (
    upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool {
            return true // For development; add proper origin checking in production
        },
    }
    
    // Store active WebSocket connections
    clients    = make(map[*websocket.Conn]bool)
    clientsMux sync.Mutex
)

func main() {
    // Initialize random seed 
    rand.Seed(time.Now().UnixNano())

    r := mux.NewRouter()

    // Serve static files (HTML, CSS, JS)
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
    
    // API endpoints
    r.HandleFunc("/ws", handleWebSocket)
    r.HandleFunc("/api/options/{symbol}", getOptionsChain)
    
    // Serve the main page
    r.HandleFunc("/", serveHome)

    // Start the streaming simulation in the background
    go simulateDataStream()

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "web/static/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    defer conn.Close()

    clientsMux.Lock()
    clients[conn] = true
    clientsMux.Unlock()

    // Remove client when connection closes
    defer func() {
        clientsMux.Lock()
        delete(clients, conn)
        clientsMux.Unlock()
    }()

    // Keep connection alive and handle messages
    for {
        messageType, _, err := conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }
        if messageType == websocket.PingMessage {
            if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
                log.Printf("Error sending pong: %v", err)
                break
            }
        }
    }
}

func getOptionsChain(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    symbol := vars["symbol"]

    // In a real application, you would fetch this data from Schwab's API
    // This is just example data
    chain := OptionChain{
        Symbol:     symbol,
        Underlying: 150.25,
        Updated:    time.Now(),
        Calls:      generateSampleOptions("call"),
        Puts:       generateSampleOptions("put"),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(chain)
}

// simulateDataStream simulates real-time data updates
func simulateDataStream() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        // Create sample data
        chain := OptionChain{
            Symbol:     "AAPL",
            Underlying: 150.25 + (rand.Float64() * 2 - 1),
            Updated:    time.Now(),
            Calls:      generateSampleOptions("call"),
            Puts:       generateSampleOptions("put"),
        }

        // Broadcast to all connected clients
        clientsMux.Lock()
        for client := range clients {
            err := client.WriteJSON(chain)
            if err != nil {
                log.Printf("WebSocket write error: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
        clientsMux.Unlock()
    }
}

func generateSampleOptions(optionType string) []OptionData {
    // Generate sample option data
    // In a real application, this would come from Schwab's API
    options := make([]OptionData, 5)
    baseStrike := 150.0

    for i := range options {
        strike := baseStrike + (float64(i-2) * 2.5)
        options[i] = OptionData{
            Strike:     strike,
            Expiration: time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
            Type:       optionType,
            Bid:        rand.Float64() * 5,
            Ask:        rand.Float64() * 5 + 0.15,
            LastPrice:  rand.Float64() * 5 + 0.10,
            Volume:     int(rand.Float64() * 1000),
            OpenInt:    int(rand.Float64() * 5000),
            Delta:      rand.Float64(),
            Gamma:      rand.Float64() * 0.1,
            Theta:      -rand.Float64(),
            Vega:       rand.Float64() * 0.2,
            ImpliedVol: rand.Float64() * 0.5,
        }
    }
    return options
}
