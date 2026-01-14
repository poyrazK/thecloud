// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Function represents a serverless function.
// Functions are event-driven units of execution (FaaS).
type Function struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Runtime   string    `json:"runtime"`   // e.g. "python3.9", "go1.21"
	Handler   string    `json:"handler"`   // Entry point (e.g. "main.Handle")
	CodePath  string    `json:"code_path"` // Path to code artifact
	Timeout   int       `json:"timeout"`   // Execution timeout in seconds
	MemoryMB  int       `json:"memory_mb"` // Memory allocation
	Status    string    `json:"status"`    // e.g. "DEPLOYING", "READY"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Invocation represents a single execution of a Function.
type Invocation struct {
	ID         uuid.UUID  `json:"id"`
	FunctionID uuid.UUID  `json:"function_id"`
	Status     string     `json:"status"` // "PENDING", "RUNNING", "SUCCESS", "FAILED"
	StartedAt  time.Time  `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	DurationMs int        `json:"duration_ms"` // Execution time in milliseconds
	StatusCode int        `json:"status_code"` // Exit code or HTTP status
	Logs       string     `json:"logs"`        // Captured stdout/stderr
}
