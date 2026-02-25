package models

import "time"

type Simulation struct {
	ID             string    `json:"id"`
	Description    string    `json:"description"`
	Preconditions  string    `json:"preconditions"`
	Rounds         int       `json:"rounds"`
	ShowOnlyResult bool      `json:"show_only_result"`
	Status         string    `json:"status"` // "running", "completed", "failed"
	Steps          []Step    `json:"steps"`
	FinalResult    string    `json:"final_result,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Step struct {
	Round     int       `json:"round"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type CreateSimulationRequest struct {
	Description    string `json:"description"`
	Preconditions  string `json:"preconditions"`
	Rounds         int    `json:"rounds"`
	ShowOnlyResult bool   `json:"show_only_result"`
}
