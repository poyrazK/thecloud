// Package domain defines core business entities.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// StackStatus represents the lifecycle state of a cloud formation stack.
type StackStatus string

const (
	// StackStatusCreateInProgress indicates resources are being provisioned.
	StackStatusCreateInProgress StackStatus = "CREATE_IN_PROGRESS"
	// StackStatusCreateComplete indicates the stack was successfully created.
	StackStatusCreateComplete StackStatus = "CREATE_COMPLETE"
	// StackStatusCreateFailed indicates the stack creation encountered an error.
	StackStatusCreateFailed StackStatus = "CREATE_FAILED"
	// StackStatusDeleteInProgress indicates the stack and its resources are being removed.
	StackStatusDeleteInProgress StackStatus = "DELETE_IN_PROGRESS"
	// StackStatusDeleteComplete indicates the stack was successfully deleted.
	StackStatusDeleteComplete StackStatus = "DELETE_COMPLETE"
	// StackStatusDeleteFailed indicates stack deletion failed.
	StackStatusDeleteFailed StackStatus = "DELETE_FAILED"
	// StackStatusRollbackInProgress indicates the stack is reconciling after a failure.
	StackStatusRollbackInProgress StackStatus = "ROLLBACK_IN_PROGRESS"
	// StackStatusRollbackComplete indicates the stack has reverted to a stable state.
	StackStatusRollbackComplete StackStatus = "ROLLBACK_COMPLETE"
	// StackStatusRollbackFailed indicates rollback failed.
	StackStatusRollbackFailed StackStatus = "ROLLBACK_FAILED"
)

// Stack represents a collection of resources defined by an Infrastructure-as-Code template.
type Stack struct {
	ID           uuid.UUID       `json:"id"`
	UserID       uuid.UUID       `json:"user_id"`
	Name         string          `json:"name"`
	Template     string          `json:"template"` // Raw YAML or JSON
	Parameters   json.RawMessage `json:"parameters" swaggertype:"string"`
	Status       StackStatus     `json:"status"`
	StatusReason string          `json:"status_reason,omitempty"`
	Resources    []StackResource `json:"resources,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// StackResource maps a logical resource from a template to a physical resource ID.
type StackResource struct {
	ID           uuid.UUID `json:"id"`
	StackID      uuid.UUID `json:"stack_id"`
	LogicalID    string    `json:"logical_id"`    // ID in template (e.g. "MyDatabase")
	PhysicalID   string    `json:"physical_id"`   // Real ID in system (e.g. UUID)
	ResourceType string    `json:"resource_type"` // e.g. "AWS::EC2::Instance" or internal type
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// TemplateValidateResponse contains the result of a template validation check.
type TemplateValidateResponse struct {
	Valid      bool     `json:"valid"`
	Errors     []string `json:"errors,omitempty"`
	Parameters []string `json:"parameters,omitempty"` // Derived parameters needed
}
