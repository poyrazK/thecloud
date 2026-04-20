// Package domain defines core business entities.
package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EnvVar represents a key-value environment variable for a function.
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FunctionUpdate describes fields that can be changed on an existing function.
type FunctionUpdate struct {
	Handler  *string   `json:"handler,omitempty"`
	Timeout  *int      `json:"timeout,omitempty"`
	MemoryMB *int      `json:"memory_mb,omitempty"`
	Status   *string   `json:"status,omitempty"`
	EnvVars  []*EnvVar `json:"env_vars,omitempty"`
}

// SetColumns returns the SQL SET clauses and argument values for a partial UPDATE.
// Only non-nil fields are included.
func (u *FunctionUpdate) SetColumns() (string, []any) {
	var cols []string
	var args []any
	argIdx := 1

	if u.Handler != nil {
		cols = append(cols, fmt.Sprintf("handler = $%d", argIdx))
		args = append(args, *u.Handler)
		argIdx++
	}
	if u.Timeout != nil {
		cols = append(cols, fmt.Sprintf("timeout_seconds = $%d", argIdx))
		args = append(args, *u.Timeout)
		argIdx++
	}
	if u.MemoryMB != nil {
		cols = append(cols, fmt.Sprintf("memory_mb = $%d", argIdx))
		args = append(args, *u.MemoryMB)
		argIdx++
	}
	if u.Status != nil {
		cols = append(cols, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *u.Status)
		argIdx++
	}
	if u.EnvVars != nil {
		cols = append(cols, fmt.Sprintf("env_vars = $%d", argIdx))
		args = append(args, u.EnvVars)
		argIdx++
	}

	return strings.Join(cols, ", "), args
}

// Function represents a serverless function.
// Functions are event-driven units of execution (FaaS).
type Function struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Name      string     `json:"name"`
	Runtime   string     `json:"runtime"`   // e.g. "python3.9", "go1.21"
	Handler   string     `json:"handler"`   // Entry point (e.g. "main.Handle")
	CodePath  string     `json:"code_path"` // Path to code artifact
	Timeout   int        `json:"timeout"`   // Execution timeout in seconds
	MemoryMB  int        `json:"memory_mb"` // Memory allocation
	Status    string     `json:"status"`    // e.g. "DEPLOYING", "READY"
	EnvVars   []*EnvVar  `json:"env_vars"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
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
