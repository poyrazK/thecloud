package services

import (
	"context"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type instanceTypeService struct {
	repo    ports.InstanceTypeRepository
	rbacSvc ports.RBACService
}

// NewInstanceTypeService creates a new InstanceTypeService.
func NewInstanceTypeService(repo ports.InstanceTypeRepository, rbacSvc ports.RBACService) ports.InstanceTypeService {
	return &instanceTypeService{
		repo:    repo,
		rbacSvc: rbacSvc,
	}
}

func (s *instanceTypeService) List(ctx context.Context) ([]*domain.InstanceType, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if userID != uuid.Nil {
		if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceRead, "*"); err != nil {
			return nil, err
		}
	}

	return s.repo.List(ctx)
}
