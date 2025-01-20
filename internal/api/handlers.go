package api

import (
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    "github.com/ryanhamamura/options-chain-go/internal/models"
    "github.com/ryanhamamura/options-chain-go/internal/stream"
)

type Handler struct {
    wsManager *stream.Manager
}

func NewHandler(wsManager *stream.Manager) *Handler {
    return &Handler{
        wsManager: wsManager,
    }
}

// ServeHome handles the main page request
func (h *Handler) ServeHome(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "web/static/index.html")
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := stream.Upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }
    defer conn.Close()

    h.wsManager.AddClient(conn)
    defer h.wsManager.RemoveClient(conn)

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

// GetOptionsChain handles requests for options chain data
func (h *Handler) GetOptionsChain(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    symbol := vars["symbol"]

    // In a real application, you would fetch this data from Schwab's API
    chain := models.OptionChain{
        Symbol:     symbol,
        Underlying: 150.25,
        Updated:    time.Now(),
        Calls:      stream.GenerateSampleOptions("call"),
        Puts:       stream.GenerateSampleOptions("put"),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(chain)
}
