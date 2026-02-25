package simulation

import (
	"fmt"
	"strings"

	"simarena/internal/llm"
	"simarena/internal/models"
)

// BuildRoundMessages constructs the chat messages for a given round.
func BuildRoundMessages(sim *models.Simulation, round int) []llm.ChatMessage {
	var system strings.Builder
	system.WriteString("You are a simulation agent. You are participating in a simulation with the following scenario and conditions.\n\n")
	system.WriteString(fmt.Sprintf("Scenario: %s\n", sim.Description))
	system.WriteString(fmt.Sprintf("Preconditions: %s\n\n", sim.Preconditions))
	system.WriteString(fmt.Sprintf("This simulation has %d rounds. You are currently on round %d/%d.\n\n", sim.Rounds, round, sim.Rounds))
	system.WriteString("Your task: analyze the situation, make decisions, and take actions within the simulation. Think strategically. Each round builds on the previous one.\n")

	if round > 1 && len(sim.Steps) > 0 {
		system.WriteString("\nHere is what happened in previous rounds:\n")
		for _, step := range sim.Steps {
			if step.Round < round {
				system.WriteString(fmt.Sprintf("\n--- Round %d ---\n%s\n", step.Round, step.Content))
			}
		}
	}

	messages := []llm.ChatMessage{
		{Role: "system", Content: system.String()},
		{Role: "user", Content: fmt.Sprintf("Execute round %d. Describe your analysis, decisions, and actions for this round. Be detailed and strategic.", round)},
	}
	return messages
}

// BuildSummaryMessages constructs the chat messages to generate a final summary.
func BuildSummaryMessages(sim *models.Simulation) []llm.ChatMessage {
	var system strings.Builder
	system.WriteString("You are a simulation agent. You completed a simulation with the following scenario:\n\n")
	system.WriteString(fmt.Sprintf("Scenario: %s\n", sim.Description))
	system.WriteString(fmt.Sprintf("Preconditions: %s\n\n", sim.Preconditions))
	system.WriteString("Here is a log of all rounds:\n")
	for _, step := range sim.Steps {
		system.WriteString(fmt.Sprintf("\n--- Round %d ---\n%s\n", step.Round, step.Content))
	}

	messages := []llm.ChatMessage{
		{Role: "system", Content: system.String()},
		{Role: "user", Content: "Provide a final summary of the simulation. What was the overall outcome? What were the key decisions and turning points? What was the final state?"},
	}
	return messages
}
