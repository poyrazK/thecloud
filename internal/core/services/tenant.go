// Package services implements core business logic.
package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// TenantService manages tenants, membership, and quota checks.
type TenantService struct {
	repo     ports.TenantRepository
	userRepo ports.UserRepository
	logger   *slog.Logger
}

// NewTenantService constructs a TenantService.
func NewTenantService(repo ports.TenantRepository, userRepo ports.UserRepository, logger *slog.Logger) *TenantService {
	return &TenantService{
		repo:     repo,
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *TenantService) CreateTenant(ctx context.Context, name, slug string, ownerID uuid.UUID) (*domain.Tenant, error) {
	// 1. Check if slug exists
	existing, _ := s.repo.GetBySlug(ctx, slug)
	if existing != nil {
		return nil, errors.New(errors.Conflict, "tenant slug already exists")
	}

	tenant := &domain.Tenant{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		OwnerID:   ownerID,
		Plan:      "free",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 2. Create tenant
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	// 3. Add owner as first member
	if err := s.repo.AddMember(ctx, tenant.ID, ownerID, "owner"); err != nil {
		return nil, err
	}

	// 4. Set default quotas
	quota := &domain.TenantQuota{
		TenantID:     tenant.ID,
		MaxInstances: 10,
		MaxVPCs:      2,
		MaxStorageGB: 50,
		MaxMemoryGB:  16,
		MaxVCPUs:     8,
	}
	if err := s.repo.UpdateQuota(ctx, quota); err != nil {
		// Log but don't fail tenant creation
		s.logger.Error("failed to set default quota", "tenant_id", tenant.ID, "err", err)
	}

	// 5. Update user's default tenant if not set
	user, err := s.userRepo.GetByID(ctx, ownerID)
	if err == nil && user.DefaultTenantID == nil {
		user.DefaultTenantID = &tenant.ID
		if err := s.userRepo.Update(ctx, user); err != nil {
			s.logger.Error("failed to set default tenant for user", "user_id", ownerID, "tenant_id", tenant.ID, "error", err)
			// Decide if this should rollback user creation.
			// Currently opting to log and continue as it is a non-critical preference setting.
		}
	}

	return tenant, nil
}

func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TenantService) InviteMember(ctx context.Context, tenantID uuid.UUID, email, role string) error {
	// In a real system, this would send an invitation email.
	// For now, we'll try to find the user by email and add them directly if they exist.
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New(errors.NotFound, "user not found")
	}

	// Check if already a member
	membership, _ := s.repo.GetMembership(ctx, tenantID, user.ID)
	if membership != nil {
		return errors.New(errors.Conflict, "user is already a member of this tenant")
	}

	return s.repo.AddMember(ctx, tenantID, user.ID, role)
}

func (s *TenantService) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	tenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}

	if tenant.OwnerID == userID {
		return errors.New(errors.Forbidden, "cannot remove tenant owner")
	}

	return s.repo.RemoveMember(ctx, tenantID, userID)
}

func (s *TenantService) SwitchTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	// Verify membership
	membership, err := s.repo.GetMembership(ctx, tenantID, userID)
	if err != nil || membership == nil {
		return errors.New(errors.Forbidden, "not a member of this tenant")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.DefaultTenantID = &tenantID
	return s.userRepo.Update(ctx, user)
}

func (s *TenantService) CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, requested int) error {
	quota, err := s.repo.GetQuota(ctx, tenantID)
	if err != nil {
		// If no quota defined, assume unlimited or default? For now, fail safe.
		// Or creating a default quota on the fly?
		// Let's assuming GetQuota returns not found error if no specific quota.
		// Ideally we should have defaults.
		return errors.Wrap(errors.Internal, "failed to get tenant quota", err)
	}

	var current, limit int
	switch resource {
	case "instances":
		current, limit = quota.UsedInstances, quota.MaxInstances
	case "vpcs":
		current, limit = quota.UsedVPCs, quota.MaxVPCs
	case "storage":
		current, limit = quota.UsedStorageGB, quota.MaxStorageGB
	case "memory":
		current, limit = quota.UsedMemoryGB, quota.MaxMemoryGB
	case "vcpus":
		current, limit = quota.UsedVCPUs, quota.MaxVCPUs
	default:
		return errors.New(errors.InvalidInput, "unknown resource type for quota check: "+resource)
	}

	if current+requested > limit {
		return errors.New(errors.QuotaExceeded, "quota exceeded for "+resource)
	}
	return nil
}

func (s *TenantService) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error) {
	return s.repo.GetMembership(ctx, tenantID, userID)
}
