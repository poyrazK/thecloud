package errors

import (
	stdlib_errors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorError(t *testing.T) {
	err := Error{
		Type:    NotFound,
		Message: "resource not found",
	}
	assert.Equal(t, "NOT_FOUND: resource not found", err.Error())

	errWithCause := Error{
		Type:    Internal,
		Message: "unexpected error",
		Cause:   fmt.Errorf("db connection failed"),
	}
	assert.Equal(t, "INTERNAL: unexpected error (cause: db connection failed)", errWithCause.Error())
}

func TestErrorUnwrap(t *testing.T) {
	cause := fmt.Errorf("db error")
	err := Wrap(Internal, "wrap error", cause)

	var e Error
	ok := stdlib_errors.As(err, &e)
	assert.True(t, ok)
	assert.Equal(t, cause, e.Unwrap())
}

func TestNew(t *testing.T) {
	err := New(InvalidInput, "invalid name")
	var e Error
	ok := stdlib_errors.As(err, &e)
	assert.True(t, ok)
	assert.Equal(t, InvalidInput, e.Type)
	assert.Equal(t, "invalid name", e.Message)
	assert.Equal(t, "INVALID_INPUT", e.Code)
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("some error")
	err := Wrap(Forbidden, "forbidden access", cause)
	var e Error
	ok := stdlib_errors.As(err, &e)
	assert.True(t, ok)
	assert.Equal(t, Forbidden, e.Type)
	assert.Equal(t, "forbidden access", e.Message)
	assert.Equal(t, cause, e.Cause)
}

func TestIs(t *testing.T) {
	err := New(Conflict, "conflict")
	require.True(t, Is(err, Conflict))
	assert.False(t, Is(fmt.Errorf("regular error"), Conflict))
}

func TestGetCause(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := Wrap(Internal, "msg", cause)
	assert.Equal(t, cause, GetCause(err))
	require.NoError(t, GetCause(fmt.Errorf("regular error")))
}
