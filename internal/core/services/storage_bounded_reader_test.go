package services

import (
	"io"
	"testing"

	apperrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoundedReaderUnderLimit(t *testing.T) {
	// 5 bytes of data, limit is 10 — should succeed without delete
	r := &countingReader{data: []byte("hello")}
	deleteCalled := false
	br := &boundedReader{
		r:      r,
		limit:  10,
		delete: func() { deleteCalled = true },
	}

	buf := make([]byte, 100)
	n, err := br.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(buf[:n]))
	assert.False(t, deleteCalled, "delete should not be called when under limit")
}

func TestBoundedReaderAtLimit(t *testing.T) {
	// 5 bytes of data, limit is exactly 5 — should succeed with EOF
	r := &countingReader{data: []byte("hello")}
	deleteCalled := false
	br := &boundedReader{
		r:      r,
		limit:  5,
		delete: func() { deleteCalled = true },
	}

	buf := make([]byte, 100)
	n, err := br.Read(buf)
	require.NoError(t, err) // EOF at exactly limit is fine
	assert.Equal(t, 5, n)
	assert.False(t, deleteCalled)
}

func TestBoundedReaderExceedsLimit(t *testing.T) {
	// 10 bytes of data, limit is 5 — second read should error ObjectTooLarge
	r := &countingReader{data: make([]byte, 10)}
	deleteCalled := false
	br := &boundedReader{
		r:      r,
		limit:  5,
		delete: func() { deleteCalled = true },
	}

	buf := make([]byte, 100)
	n1, err1 := br.Read(buf)
	require.NoError(t, err1)
	assert.Equal(t, 5, n1) // first 5 bytes allowed

	n2, err2 := br.Read(buf)
	assert.True(t, deleteCalled, "delete must be called on oversize")
	assert.Equal(t, 0, n2)
	var appErr1 apperrors.Error
	assert.ErrorAs(t, err2, &appErr1)
	assert.Equal(t, apperrors.ObjectTooLarge, appErr1.Type)
}

func TestBoundedReaderExceedsLimitInSingleRead(t *testing.T) {
	// Simulates: io.LimitReaderunderlying, 10) wraps a 10-byte reader.
	// io.LimitReader respects buffer size (only gives us at most 'remaining' bytes
	// per call) but tracks consumed bytes internally. Here we test the case
	// where the underlying Read returns exactly 'remaining' bytes and advances
	// count to the limit — the next Read must error.
	r := &countingReader{data: make([]byte, 10)}
	br := &boundedReader{
		r:      r,
		limit:  5,
		delete: func() {},
	}

	buf := make([]byte, 100)
	n1, err1 := br.Read(buf)
	require.NoError(t, err1)
	assert.Equal(t, 5, n1) // count=5, exactly at limit

	// Read again — count == limit, must return ObjectTooLarge
	n2, err2 := br.Read(buf)
	assert.Equal(t, 0, n2)
	var appErr2 apperrors.Error
	assert.ErrorAs(t, err2, &appErr2)
	assert.Equal(t, apperrors.ObjectTooLarge, appErr2.Type)
}

// countingReader returns data byte-by-byte and tracks position.
type countingReader struct {
	data []byte
	pos  int
}

func (c *countingReader) Read(p []byte) (n int, err error) {
	if c.pos >= len(c.data) {
		return 0, io.EOF
	}
	n = copy(p, c.data[c.pos:])
	c.pos += n
	return n, nil
}
