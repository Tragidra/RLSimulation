package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"simarena/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins for local dev
	},
}

// Hub manages WebSocket connections grouped by simulation ID.
type Hub struct {
	mu    sync.RWMutex
	conns map[string]map[*websocket.Conn]bool
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		conns: make(map[string]map[*websocket.Conn]bool),
	}
}

// Register adds a connection for a simulation.
func (h *Hub) Register(simID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[simID] == nil {
		h.conns[simID] = make(map[*websocket.Conn]bool)
	}
	h.conns[simID][conn] = true
}

// Unregister removes a connection.
func (h *Hub) Unregister(simID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.conns[simID]; ok {
		delete(clients, conn)
		if len(clients) == 0 {
			delete(h.conns, simID)
		}
	}
}

// BroadcastStep sends a step to all connected clients for a simulation.
func (h *Hub) BroadcastStep(simID string, step models.Step) {
	h.mu.RLock()
	clients := h.conns[simID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	data, err := json.Marshal(step)
	if err != nil {
		log.Printf("ERROR: marshal step for broadcast: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ERROR: write to websocket: %v", err)
			conn.Close()
			delete(h.conns[simID], conn)
		}
	}
}
