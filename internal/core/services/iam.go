package services

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type iamService struct {
	repo     ports.IAMRepository
	auditSvc ports.AuditService
	eventSvc ports.EventService
	logger   *slog.Logger
}

// NewIAMService creates a new IAM service.
func NewIAMService(repo ports.IAMRepository, auditSvc ports.AuditService, eventSvc ports.EventService, logger *slog.Logger) *iamService {
	return &iamService{
		repo:     repo,
		auditSvc: auditSvc,
		eventSvc: eventSvc,
		logger:   logger,
	}
}

func (s *iamService) CreatePolicy(ctx context.Context, policy *domain.Policy) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	policy.TenantID = tenantID

	if err := s.repo.CreatePolicy(ctx, tenantID, policy); err != nil {
		return err
	}

	_ = s.eventSvc.RecordEvent(ctx, "IAM_POLICY_CREATE", policy.ID.String(), "POLICY", map[string]interface{}{"name": policy.Name})
	return nil
}

func (s *iamService) GetPolicyByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	return s.repo.GetPolicyByID(ctx, tenantID, id)
}

func (s *iamService) ListPolicies(ctx context.Context) ([]*domain.Policy, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	return s.repo.ListPolicies(ctx, tenantID)
}

func (s *iamService) UpdatePolicy(ctx context.Context, policy *domain.Policy) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	if err := s.repo.UpdatePolicy(ctx, tenantID, policy); err != nil {
		return err
	}
	_ = s.eventSvc.RecordEvent(ctx, "IAM_POLICY_UPDATE", policy.ID.String(), "POLICY", map[string]interface{}{"name": policy.Name})
	return nil
}

func (s *iamService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	if err := s.repo.DeletePolicy(ctx, tenantID, id); err != nil {
		return err
	}
	_ = s.eventSvc.RecordEvent(ctx, "IAM_POLICY_DELETE", id.String(), "POLICY", nil)
	return nil
}

func (s *iamService) AttachPolicyToUser(ctx context.Context, userID uuid.UUID, policyID uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	if err := s.repo.AttachPolicyToUser(ctx, tenantID, userID, policyID); err != nil {
		return err
	}
	_ = s.auditSvc.Log(ctx, userID, "iam.policy_attach", "user", userID.String(), map[string]interface{}{"policy_id": policyID})
	return nil
}

func (s *iamService) DetachPolicyFromUser(ctx context.Context, userID uuid.UUID, policyID uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	if err := s.repo.DetachPolicyFromUser(ctx, tenantID, userID, policyID); err != nil {
		return err
	}
	_ = s.auditSvc.Log(ctx, userID, "iam.policy_detach", "user", userID.String(), map[string]interface{}{"policy_id": policyID})
	return nil
}

func (s *iamService) GetPoliciesForUser(ctx context.Context, userID uuid.UUID) ([]*domain.Policy, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	return s.repo.GetPoliciesForUser(ctx, tenantID, userID)
}
