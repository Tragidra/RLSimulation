package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates and configures the HTTP router.
func NewRouter(h *Handler, corsOrigin string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{corsOrigin},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}))

	r.Route("/api/simulations", func(r chi.Router) {
		r.Post("/", h.CreateSimulation)
		r.Get("/", h.ListSimulations)
		r.Get("/{id}", h.GetSimulation)
		r.Delete("/{id}", h.DeleteSimulation)
		r.Get("/{id}/ws", h.WebSocketHandler)
	})

	return r
}
