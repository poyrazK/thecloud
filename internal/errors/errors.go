package errors

import (
	"fmt"
)

type Type string

const (
	NotFound      Type = "NOT_FOUND"
	InvalidInput  Type = "INVALID_INPUT"
	Internal      Type = "INTERNAL"
	Unauthorized  Type = "UNAUTHORIZED"
	Conflict      Type = "CONFLICT"
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
