package simulation

import (
	"context"
	"log"
	"time"

	"simarena/internal/llm"
	"simarena/internal/models"
	"simarena/internal/storage"
)

// StepCallback is called when a new step is completed.
type StepCallback func(simID string, step models.Step)

// Engine orchestrates simulation runs.
type Engine struct {
	llmClient *llm.Client
	store     *storage.JSONStore
	onStep    StepCallback
}

// NewEngine creates a new simulation engine.
func NewEngine(client *llm.Client, store *storage.JSONStore, onStep StepCallback) *Engine {
	return &Engine{
		llmClient: client,
		store:     store,
		onStep:    onStep,
	}
}

// Run executes a simulation asynchronously.
func (e *Engine) Run(sim *models.Simulation) {
	go e.run(sim)
}

func (e *Engine) run(sim *models.Simulation) {
	ctx := context.Background()

	for round := 1; round <= sim.Rounds; round++ {
		messages := BuildRoundMessages(sim, round)

		content, err := e.llmClient.ChatCompletionStream(ctx, messages, nil)
		if err != nil {
			log.Printf("ERROR: simulation %s round %d failed: %v", sim.ID, round, err)
			sim.Status = "failed"
			if updateErr := e.store.Update(*sim); updateErr != nil {
				log.Printf("ERROR: failed to update simulation status: %v", updateErr)
			}
			// Broadcast a failure step so WebSocket clients know
			failStep := models.Step{
				Round:     round,
				Content:   "Error: " + err.Error(),
				Timestamp: time.Now(),
			}
			if e.onStep != nil {
				e.onStep(sim.ID, failStep)
			}
			return
		}

		step := models.Step{
			Round:     round,
			Content:   content,
			Timestamp: time.Now(),
		}
		sim.Steps = append(sim.Steps, step)

		if err := e.store.Update(*sim); err != nil {
			log.Printf("ERROR: failed to save step: %v", err)
		}

		if e.onStep != nil {
			e.onStep(sim.ID, step)
		}
	}

	// Generate final summary
	summaryMessages := BuildSummaryMessages(sim)
	summary, err := e.llmClient.ChatCompletion(ctx, summaryMessages)
	if err != nil {
		log.Printf("ERROR: simulation %s summary failed: %v", sim.ID, err)
		summary = "Summary generation failed: " + err.Error()
	}

	sim.FinalResult = summary
	sim.Status = "completed"
	if err := e.store.Update(*sim); err != nil {
		log.Printf("ERROR: failed to update simulation: %v", err)
	}

	// Broadcast completion
	if e.onStep != nil {
		e.onStep(sim.ID, models.Step{
			Round:     -1, // sentinel: indicates completion
			Content:   summary,
			Timestamp: time.Now(),
		})
	}
}
