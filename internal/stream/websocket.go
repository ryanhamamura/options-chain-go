package stream

import (
    "log"
    "math/rand"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/ryanhamamura/options-chain-go/internal/models"
)

var (
    // Upgrader configuration for WebSocket connections
    Upgrader = websocket.Upgrader{
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

// Manager handles WebSocket client connections and broadcasting
type Manager struct {
    clients    map[*websocket.Conn]bool
    clientsMux sync.Mutex
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
    return &Manager{
        clients: make(map[*websocket.Conn]bool),
    }
}

// AddClient registers a new WebSocket client
func (m *Manager) AddClient(conn *websocket.Conn) {
    m.clientsMux.Lock()
    m.clients[conn] = true
    m.clientsMux.Unlock()
}

// RemoveClient removes a WebSocket client
func (m *Manager) RemoveClient(conn *websocket.Conn) {
    m.clientsMux.Lock()
    delete(m.clients, conn)
    m.clientsMux.Unlock()
}

// BroadcastOptionChain sends option chain data to all connected clients
func (m *Manager) BroadcastOptionChain(chain models.OptionChain) {
    m.clientsMux.Lock()
    for client := range m.clients {
        err := client.WriteJSON(chain)
        if err != nil {
            log.Printf("WebSocket write error: %v", err)
            client.Close()
            delete(m.clients, client)
        }
    }
    m.clientsMux.Unlock()
}

// StartSimulation begins simulating option chain updates
func (m *Manager) StartSimulation() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        chain := models.OptionChain{
            Symbol:     "AAPL",
            Underlying: 150.25 + (rand.Float64() * 2 - 1),
            Updated:    time.Now(),
            Calls:      GenerateSampleOptions("call"),
            Puts:       GenerateSampleOptions("put"),
        }
        m.BroadcastOptionChain(chain)
    }
}

// generateSampleOptions creates sample option data for testing
func GenerateSampleOptions(optionType string) []models.OptionData {
    options := make([]models.OptionData, 5)
    baseStrike := 150.0

    for i := range options {
        strike := baseStrike + (float64(i-2) * 2.5)
        options[i] = models.OptionData{
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
