package services

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIAMService(t *testing.T) {
	repo := new(mocks.IAMRepository)
	auditSvc := new(mocks.AuditService)
	eventSvc := new(mocks.EventService)
	logger := slog.Default()
	svc := NewIAMService(repo, auditSvc, eventSvc, logger)
	ctx := context.Background()

	policy := &domain.Policy{
		ID:   uuid.New(),
		Name: "Test",
	}

	t.Run("CreatePolicy", func(t *testing.T) {
		repo.On("CreatePolicy", ctx, policy).Return(nil).Once()
		eventSvc.On("RecordEvent", ctx, "IAM_POLICY_CREATE", policy.ID.String(), "POLICY", mock.Anything).Return(nil).Once()

		err := svc.CreatePolicy(ctx, policy)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("AttachPolicy", func(t *testing.T) {
		userID := uuid.New()
		repo.On("AttachPolicyToUser", ctx, userID, policy.ID).Return(nil).Once()
		auditSvc.On("Log", ctx, userID, "iam.policy_attach", "user", userID.String(), mock.Anything).Return(nil).Once()

		err := svc.AttachPolicyToUser(ctx, userID, policy.ID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
