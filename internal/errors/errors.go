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
)

type Error struct {
	Type    Type   `json:"type"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func New(t Type, msg string) error {
	return Error{Type: t, Message: msg}
}

func Wrap(t Type, msg string, err error) error {
	return Error{Type: t, Message: msg, Cause: err}
}

func Is(err error, t Type) bool {
	if e, ok := err.(Error); ok {
		return e.Type == t
	}
	return false
}
