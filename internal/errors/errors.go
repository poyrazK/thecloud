package errors

import (
	"fmt"
)

type Type string

const (
	NotFound     Type = "NOT_FOUND"
	InvalidInput Type = "INVALID_INPUT"
	Internal     Type = "INTERNAL"
	Unauthorized Type = "UNAUTHORIZED"
	Conflict     Type = "CONFLICT"

	// Storage Errors
	BucketNotFound Type = "BUCKET_NOT_FOUND"
	ObjectNotFound Type = "OBJECT_NOT_FOUND"
	ObjectTooLarge Type = "OBJECT_TOO_LARGE"

	// Networking Errors
	InvalidPortFormat  Type = "INVALID_PORT_FORMAT"
	PortConflict       Type = "PORT_CONFLICT"
	TooManyPorts       Type = "TOO_MANY_PORTS"
	InstanceNotRunning Type = "INSTANCE_NOT_RUNNING"
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

// Unwrap implements the errors.Unwrap interface for error chain support
func (e Error) Unwrap() error {
	return e.Cause
}

func New(t Type, msg string) error {
	return Error{Type: t, Message: msg, Code: string(t)}
}

func Wrap(t Type, msg string, err error) error {
	return Error{Type: t, Message: msg, Code: string(t), Cause: err}
}

func Is(err error, t Type) bool {
	if e, ok := err.(Error); ok {
		return e.Type == t
	}
	return false
}

// GetCause returns the underlying cause for logging purposes (not for client exposure)
func GetCause(err error) error {
	if e, ok := err.(Error); ok {
		return e.Cause
	}
	return nil
}
