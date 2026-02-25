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
	maxTokens := models.DepthToMaxTokens(sim.Depth)

	for round := 1; round <= sim.Rounds; round++ {
		for _, agent := range sim.Agents {
			messages := BuildAgentRoundMessages(sim, agent, round)

			content, err := e.llmClient.ChatCompletionStream(ctx, messages, maxTokens, nil)
			if err != nil {
				log.Printf("ERROR: simulation %s round %d agent %s failed: %v", sim.ID, round, agent.Name, err)
				sim.Status = "failed"
				if updateErr := e.store.Update(*sim); updateErr != nil {
					log.Printf("ERROR: failed to update simulation status: %v", updateErr)
				}
				failStep := models.Step{
					Round:     round,
					AgentID:   agent.ID,
					AgentName: agent.Name,
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
				AgentID:   agent.ID,
				AgentName: agent.Name,
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
	}

	// Generate final summary
	summaryMessages := BuildSummaryMessages(sim)
	summary, err := e.llmClient.ChatCompletion(ctx, summaryMessages, maxTokens)
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
