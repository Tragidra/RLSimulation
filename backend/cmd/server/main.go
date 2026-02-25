package main

import (
	"log"
	"net/http"
	"os"

	"simarena/internal/api"
	"simarena/internal/llm"
	"simarena/internal/models"
	"simarena/internal/simulation"
	"simarena/internal/storage"
)

func main() {
	// Configuration from env or defaults
	port := getEnv("PORT", "8080")
	corsOrigin := getEnv("CORS_ORIGIN", "http://localhost:5173")
	llmBaseURL := getEnv("LLM_BASE_URL", "http://localhost:7090/v1")
	llmModel := getEnv("LLM_MODEL", "openai/gpt-oss-20b")
	llmAPIKey := getEnv("LLM_API_KEY", "not-needed")
	dataPath := getEnv("DATA_PATH", "./data")

	// Storage
	store, err := storage.NewJSONStore(dataPath)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// LLM Client
	llmCfg := llm.DefaultConfig()
	llmCfg.BaseURL = llmBaseURL
	llmCfg.Model = llmModel
	llmCfg.APIKey = llmAPIKey
	llmClient := llm.NewClient(llmCfg)

	// WebSocket hub
	hub := api.NewHub()

	// Simulation engine
	engine := simulation.NewEngine(llmClient, store, func(simID string, step models.Step) {
		hub.BroadcastStep(simID, step)
	})

	// HTTP handler and router
	handler := api.NewHandler(store, engine, hub)
	router := api.NewRouter(handler, corsOrigin)

	log.Printf("SimArena backend starting on :%s", port)
	log.Printf("CORS origin: %s", corsOrigin)
	log.Printf("LLM endpoint: %s (model: %s)", llmBaseURL, llmModel)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
