// Package domain defines core business entities.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// EnvVar represents a key-value environment variable for a function.
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FunctionUpdate describes fields that can be updated on a function.
// All fields are pointers so nil means "do not update".
type FunctionUpdate struct {
	Handler   *string    `json:"handler,omitempty"`
	Timeout   *int       `json:"timeout,omitempty"`
	MemoryMB  *int       `json:"memory_mb,omitempty"`
	Status    string     `json:"status,omitempty"`
	EnvVars   []*EnvVar `json:"env_vars,omitempty"`
}

// SetColumns returns the names of non-zero/nil fields for dynamic SQL UPDATE.
func (u *FunctionUpdate) SetColumns() []string {
	var cols []string
	if u.Handler != nil {
		cols = append(cols, "handler")
	}
	if u.Timeout != nil {
		cols = append(cols, "timeout_seconds")
	}
	if u.MemoryMB != nil {
		cols = append(cols, "memory_mb")
	}
	if u.Status != "" {
		cols = append(cols, "status")
	}
	if u.EnvVars != nil {
		cols = append(cols, "env_vars")
	}
	return cols
}

// Function represents a serverless function.
// Functions are event-driven units of execution (FaaS).
type Function struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Runtime   string    `json:"runtime"`   // e.g. "python3.9", "go1.21"
	Handler   string    `json:"handler"`   // Entry point (e.g. "main.Handle")
	CodePath  string    `json:"code_path"` // Path to code artifact
	Timeout   int       `json:"timeout"`   // Execution timeout in seconds
	MemoryMB  int       `json:"memory_mb"` // Memory allocation
	Status    string    `json:"status"`    // e.g. "DEPLOYING", "READY"
	EnvVars   []*EnvVar `json:"env_vars,omitempty"`
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
