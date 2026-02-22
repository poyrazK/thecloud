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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type MockIAMRepository struct {
	mock.Mock
}

func (m *MockIAMRepository) CreatePolicy(ctx context.Context, tenantID uuid.UUID, policy *domain.Policy) error {
	return m.Called(ctx, tenantID, policy).Error(0)
}
func (m *MockIAMRepository) GetPolicyByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Policy, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Policy)
	return r0, args.Error(1)
}
func (m *MockIAMRepository) ListPolicies(ctx context.Context, tenantID uuid.UUID) ([]*domain.Policy, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Policy)
	return r0, args.Error(1)
}
func (m *MockIAMRepository) UpdatePolicy(ctx context.Context, tenantID uuid.UUID, policy *domain.Policy) error {
	return m.Called(ctx, tenantID, policy).Error(0)
}
func (m *MockIAMRepository) DeletePolicy(ctx context.Context, tenantID, id uuid.UUID) error {
	return m.Called(ctx, tenantID, id).Error(0)
}
func (m *MockIAMRepository) AttachPolicyToUser(ctx context.Context, tenantID, userID, policyID uuid.UUID) error {
	return m.Called(ctx, tenantID, userID, policyID).Error(0)
}
func (m *MockIAMRepository) DetachPolicyFromUser(ctx context.Context, tenantID, userID, policyID uuid.UUID) error {
	return m.Called(ctx, tenantID, userID, policyID).Error(0)
}
func (m *MockIAMRepository) GetPoliciesForUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Policy, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Policy)
	return r0, args.Error(1)
}

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
	
	t.Run("GetPoliciesForUser", func(t *testing.T) {
		userID := uuid.New()
		mockRepo.On("GetPoliciesForUser", mock.Anything, tenantID, userID).Return([]*domain.Policy{}, nil).Once()
		
		res, err := svc.GetPoliciesForUser(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})
}
