package httphandlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTraceID(t *testing.T) {
	t.Parallel()

	id := generateTraceID()
	assert.Len(t, id, 32, "trace ID should be 32 hex characters (16 bytes)")
}

func TestGenerateSpanID(t *testing.T) {
	t.Parallel()

	id := generateSpanID()
	assert.Len(t, id, 16, "span ID should be 16 hex characters (8 bytes)")
}

func TestGenerateTraceID_Uniqueness(t *testing.T) {
	t.Parallel()

	// Generate two IDs and ensure they're different
	id1 := generateTraceID()
	id2 := generateTraceID()
	assert.NotEqual(t, id1, id2, "consecutive trace IDs should be unique")
}

func TestGenerateSpanID_Uniqueness(t *testing.T) {
	t.Parallel()

	id1 := generateSpanID()
	id2 := generateSpanID()
	assert.NotEqual(t, id1, id2, "consecutive span IDs should be unique")
}
