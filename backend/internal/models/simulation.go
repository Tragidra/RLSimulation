package models

import "time"

type Agent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type Simulation struct {
	ID             string    `json:"id"`
	Description    string    `json:"description"`
	Preconditions  string    `json:"preconditions"`
	Rounds         int       `json:"rounds"`
	ShowOnlyResult bool      `json:"show_only_result"`
	Agents         []Agent   `json:"agents"`
	Language       string    `json:"language"` // "en" or "ru"
	Depth          string    `json:"depth"`    // "shallow", "medium", "deep"
	Status         string    `json:"status"`   // "running", "completed", "failed"
	Steps          []Step    `json:"steps"`
	FinalResult    string    `json:"final_result,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// IsInteractive returns true if any agent has a non-empty role.
func (s *Simulation) IsInteractive() bool {
	for _, a := range s.Agents {
		if a.Role != "" {
			return true
		}
	}
	return false
}

// DepthToMaxTokens maps depth to max_tokens (0 = no limit).
func DepthToMaxTokens(depth string) int {
	switch depth {
	case "shallow":
		return 400
	case "deep":
		return 0
	default:
		return 1200
	}
}

type Step struct {
	Round     int       `json:"round"`
	AgentID   string    `json:"agent_id"`
	AgentName string    `json:"agent_name"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateSimulationRequest struct {
	Description    string         `json:"description"`
	Preconditions  string         `json:"preconditions"`
	Rounds         int            `json:"rounds"`
	ShowOnlyResult bool           `json:"show_only_result"`
	Agents         []AgentRequest `json:"agents"`
	Language       string         `json:"language"`
	Depth          string         `json:"depth"`
}

type AgentRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
