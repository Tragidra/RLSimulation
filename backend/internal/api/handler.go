package api

import (
	"encoding/json"
	"net/http"
	"time"

	"simarena/internal/models"
	"simarena/internal/simulation"
	"simarena/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	store  *storage.JSONStore
	engine *simulation.Engine
	hub    *Hub
}

// NewHandler creates a new Handler.
func NewHandler(store *storage.JSONStore, engine *simulation.Engine, hub *Hub) *Handler {
	return &Handler{
		store:  store,
		engine: engine,
		hub:    hub,
	}
}

// CreateSimulation handles POST /api/simulations.
func (h *Handler) CreateSimulation(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, `{"error":"description is required"}`, http.StatusBadRequest)
		return
	}
	if req.Rounds < 1 || req.Rounds > 20 {
		http.Error(w, `{"error":"rounds must be between 1 and 20"}`, http.StatusBadRequest)
		return
	}

	sim := models.Simulation{
		ID:             uuid.New().String(),
		Description:    req.Description,
		Preconditions:  req.Preconditions,
		Rounds:         req.Rounds,
		ShowOnlyResult: req.ShowOnlyResult,
		Status:         "running",
		Steps:          []models.Step{},
		CreatedAt:      time.Now(),
	}

	if err := h.store.Create(sim); err != nil {
		http.Error(w, `{"error":"failed to save simulation"}`, http.StatusInternalServerError)
		return
	}

	h.engine.Run(&sim)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sim)
}

// ListSimulations handles GET /api/simulations.
func (h *Handler) ListSimulations(w http.ResponseWriter, r *http.Request) {
	sims, err := h.store.List()
	if err != nil {
		http.Error(w, `{"error":"failed to list simulations"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sims)
}

// GetSimulation handles GET /api/simulations/{id}.
func (h *Handler) GetSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sim, err := h.store.Get(id)
	if err != nil {
		http.Error(w, `{"error":"simulation not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sim)
}

// DeleteSimulation handles DELETE /api/simulations/{id}.
func (h *Handler) DeleteSimulation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.store.Delete(id); err != nil {
		http.Error(w, `{"error":"simulation not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WebSocketHandler handles WS /api/simulations/{id}/ws.
func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Verify simulation exists
	_, err := h.store.Get(id)
	if err != nil {
		http.Error(w, `{"error":"simulation not found"}`, http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.Register(id, conn)

	// Keep connection alive and handle close
	go func() {
		defer func() {
			h.hub.Unregister(id, conn)
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
