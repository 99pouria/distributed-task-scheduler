package task

import (
	"encoding/json"
	"time"
)

// Priority defines the task priority levels.
type Priority int

const (
	Low Priority = iota
	Medium
	High
)

func (p Priority) String() string {
	return []string{"Low", "Medium", "High"}[p]
}

// Status is state of a task.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Task struct {
	ID        string          `json:"id"`
	Priority  Priority        `json:"priority"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
	Status    Status          `json:"status"`
}
