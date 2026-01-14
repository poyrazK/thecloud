package simpleaudit

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAuditLoggerLog(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	audit := NewSimpleAuditLogger(logger)

	userID := uuid.New()
	entry := &domain.AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       "CREATE",
		ResourceType: "instance",
		ResourceID:   "123",
		IPAddress:    testutil.TestIPLocalhost,
		UserAgent:    testutil.TestUserAgent,
		Details:      map[string]interface{}{"foo": "bar"},
		CreatedAt:    time.Now(),
	}

	err := audit.Log(context.Background(), entry)
	assert.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "AUDIT_LOG")
	assert.Contains(t, logOutput, "security_audit")
	assert.Contains(t, logOutput, "CREATE")
	assert.Contains(t, logOutput, userID.String())
	assert.Contains(t, logOutput, "instance")
	assert.Contains(t, logOutput, "123")
	assert.Contains(t, logOutput, testutil.TestIPLocalhost)
	assert.Contains(t, logOutput, testutil.TestUserAgent)
	assert.Contains(t, logOutput, "foo")
	assert.Contains(t, logOutput, "bar")
}

func TestSimpleAuditLoggerLogAnonymous(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	audit := NewSimpleAuditLogger(logger)

	entry := &domain.AuditLog{
		ID:           uuid.New(),
		UserID:       uuid.Nil,
		Action:       "LOGIN",
		ResourceType: "auth",
		CreatedAt:    time.Now(),
	}

	err := audit.Log(context.Background(), entry)
	assert.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "anonymous")
}

func TestSimpleAuditLoggerLogEmptyDetails(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	audit := NewSimpleAuditLogger(logger)

	entry := &domain.AuditLog{
		ID:           uuid.New(),
		Action:       "TEST",
		ResourceType: "test",
		Details:      map[string]interface{}{},
	}

	err := audit.Log(context.Background(), entry)
	assert.NoError(t, err)

	logOutput := buf.String()
	// slog with JSON handler handles empty map gracefully
	assert.Contains(t, logOutput, "details")
}
