package k8s

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockInstanceService is already declared in health_test.go, so we reuse it here.

func TestServiceExecutor(t *testing.T) {
	instSvc := new(mockInstanceService)
	instID := uuid.New()
	exec := NewServiceExecutor(instSvc, instID)
	ctx := context.Background()

	t.Run("Run", func(t *testing.T) {
		expectedCmd := []string{"/bin/sh", "-c", "echo hello"}
		instSvc.On("Exec", ctx, instID.String(), expectedCmd).Return("hello", nil).Once()

		out, err := exec.Run(ctx, "echo hello")
		assert.NoError(t, err)
		assert.Equal(t, "hello", out)
	})

	t.Run("WaitForReady Success", func(t *testing.T) {
		expectedCmd := []string{"/bin/sh", "-c", "echo ready"}
		instSvc.On("Exec", mock.Anything, instID.String(), expectedCmd).Return("ready", nil).Once()

		err := exec.WaitForReady(ctx, 5*time.Second)
		assert.NoError(t, err)
	})

	t.Run("WaitForReady Timeout", func(t *testing.T) {
		instSvc.On("Exec", mock.Anything, instID.String(), mock.Anything).Return("", fmt.Errorf("not yet")).Maybe()

		err := exec.WaitForReady(ctx, 100*time.Millisecond) // Short timeout for test
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})
}
