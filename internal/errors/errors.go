// Package errors provides custom error types and utilities for the application.
package errors

import (
	stdlib_errors "errors"
	"fmt"
)

// Type identifies a category of application error.
type Type string

// Common application error types.
const (
	NotFound              Type = "NOT_FOUND"
	InvalidInput          Type = "INVALID_INPUT"
	Internal              Type = "INTERNAL"
	Unauthorized          Type = "UNAUTHORIZED"
	Conflict              Type = "CONFLICT"
	Forbidden             Type = "FORBIDDEN"
	PermissionDenied      Type = "PERMISSION_DENIED"
	ResourceLimitExceeded Type = "RESOURCE_LIMIT_EXCEEDED"
	QuotaExceeded         Type = "QUOTA_EXCEEDED"
	NotImplemented        Type = "NOT_IMPLEMENTED"

	// Storage Errors
	BucketNotFound Type = "BUCKET_NOT_FOUND"
	ObjectNotFound Type = "OBJECT_NOT_FOUND"
	ObjectTooLarge Type = "OBJECT_TOO_LARGE"

	// Networking Errors
	InvalidPortFormat  Type = "INVALID_PORT_FORMAT"
	PortConflict       Type = "PORT_CONFLICT"
	TooManyPorts       Type = "TOO_MANY_PORTS"
	InstanceNotRunning Type = "INSTANCE_NOT_RUNNING"

	// Load Balancer Errors
	LBNotFound     Type = "LB_NOT_FOUND"
	LBTargetExists Type = "LB_TARGET_EXISTS"
	LBCrossVPC     Type = "LB_CROSS_VPC"
)

// Error represents an API error that can be safely returned to clients.
// The Cause field is intentionally omitted from JSON to prevent internal details from leaking.
type Error struct {
	Type    Type   `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"` // Optional error code for programmatic handling
	Cause   error  `json:"-"`              // Never exposed to clients
}

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap implements the stdlib_errors.Unwrap interface for error chain support
func (e Error) Unwrap() error {
	return e.Cause
}

// New creates an Error with a type and message.
func New(t Type, msg string) error {
	return Error{Type: t, Message: msg, Code: string(t)}
}

// Wrap creates an Error with a cause.
func Wrap(t Type, msg string, err error) error {
	return Error{Type: t, Message: msg, Code: string(t), Cause: err}
}

// Is checks whether an error is of the given Type.
func Is(err error, t Type) bool {
	var e Error
	if stdlib_errors.As(err, &e) {
		return e.Type == t
	}
	return false
}

// As finds the first error in err's chain that matches target.
func As(err error, target any) bool {
	return stdlib_errors.As(err, target)
}

// GetCause returns the underlying cause for logging purposes (not for client exposure)
func GetCause(err error) error {
	var e Error
	if stdlib_errors.As(err, &e) {
		return e.Cause
	}
	return nil
}

// Convenience error values for common load balancer failures.
var (
	ErrLBNotFound     = New(LBNotFound, "load balancer not found")
	ErrLBTargetExists = New(LBTargetExists, "target already registered")
	ErrLBCrossVPC     = New(LBCrossVPC, "target must be in same VPC as LB")
)
