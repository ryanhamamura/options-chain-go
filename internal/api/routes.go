package api

import (
    "net/http"

    "github.com/gorilla/mux"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(r *mux.Router, h *Handler) {
    // Serve static files
    r.PathPrefix("/static/").Handler(
        http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
    
    // API endpoints
    r.HandleFunc("/ws", h.HandleWebSocket)
    r.HandleFunc("/api/options/{symbol}", h.GetOptionsChain)
    
    // Serve the main page
    r.HandleFunc("/", h.ServeHome)
}
