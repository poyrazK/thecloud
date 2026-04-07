package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIAMService_Unit(t *testing.T) {
	mockRepo := new(MockIAMRepository)
	mockAuditSvc := new(MockAuditService)
	mockEventSvc := new(MockEventService)
	svc := services.NewIAMService(mockRepo, mockAuditSvc, mockEventSvc, slog.Default())

	ctx := context.Background()
	tenantID := uuid.New()
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("CreatePolicy", func(t *testing.T) {
		policy := &domain.Policy{Name: "test-policy"}
		mockRepo.On("CreatePolicy", mock.Anything, tenantID, mock.Anything).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "IAM_POLICY_CREATE", mock.Anything, "POLICY", mock.Anything).Return(nil).Once()

		err := svc.CreatePolicy(ctx, policy)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("AttachPolicyToUser", func(t *testing.T) {
		userID := uuid.New()
		policyID := uuid.New()
		mockRepo.On("AttachPolicyToUser", mock.Anything, tenantID, userID, policyID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "iam.policy_attach", "user", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.AttachPolicyToUser(ctx, userID, policyID)
		require.NoError(t, err)
	})

	t.Run("DetachPolicyFromUser", func(t *testing.T) {
		userID := uuid.New()
		policyID := uuid.New()
		mockRepo.On("DetachPolicyFromUser", mock.Anything, tenantID, userID, policyID).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "iam.policy_detach", "user", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.DetachPolicyFromUser(ctx, userID, policyID)
		require.NoError(t, err)
	})

	t.Run("GetPoliciesForUser", func(t *testing.T) {
		userID := uuid.New()
		mockRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{}, nil).Once()

		res, err := svc.GetPoliciesForUser(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("UpdatePolicy", func(t *testing.T) {
		policy := &domain.Policy{ID: uuid.New(), Name: "updated"}
		mockRepo.On("UpdatePolicy", mock.Anything, tenantID, policy).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "IAM_POLICY_UPDATE", policy.ID.String(), "POLICY", mock.Anything).Return(nil).Once()

		err := svc.UpdatePolicy(ctx, policy)
		require.NoError(t, err)
	})

	t.Run("DeletePolicy", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("DeletePolicy", mock.Anything, tenantID, id).Return(nil).Once()
		mockEventSvc.On("RecordEvent", mock.Anything, "IAM_POLICY_DELETE", id.String(), "POLICY", mock.Anything).Return(nil).Once()

		err := svc.DeletePolicy(ctx, id)
		require.NoError(t, err)
	})
}
