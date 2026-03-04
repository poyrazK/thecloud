// Package services implements core business workflows.
package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// ContainerService manages deployments and their containers.
type ContainerService struct {
	repo     ports.ContainerRepository
	rbacSvc  ports.RBACService
	eventSvc ports.EventService
	auditSvc ports.AuditService
}

// NewContainerService constructs a ContainerService with its dependencies.
func NewContainerService(repo ports.ContainerRepository, rbacSvc ports.RBACService, eventSvc ports.EventService, auditSvc ports.AuditService) ports.ContainerService {
	return &ContainerService{
		repo:     repo,
		rbacSvc:  rbacSvc,
		eventSvc: eventSvc,
		auditSvc: auditSvc,
	}
}

func (s *ContainerService) CreateDeployment(ctx context.Context, name, image string, replicas int, ports string) (*domain.Deployment, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceLaunch); err != nil {
		return nil, err
	}

	dep := &domain.Deployment{
		ID:           uuid.New(),
		UserID:       userID,
		TenantID:     tenantID,
		Name:         name,
		Image:        image,
		Replicas:     replicas,
		CurrentCount: 0,
		Ports:        ports,
		Status:       domain.DeploymentStatusScaling,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateDeployment(ctx, dep); err != nil {
		return nil, err
	}

	_ = s.eventSvc.RecordEvent(ctx, "DEPLOYMENT_CREATED", dep.ID.String(), "DEPLOYMENT", nil)

	_ = s.auditSvc.Log(ctx, dep.UserID, "container.deployment_create", "deployment", dep.ID.String(), map[string]interface{}{
		"name":  dep.Name,
		"image": dep.Image,
	})

	return dep, nil
}

func (s *ContainerService) ListDeployments(ctx context.Context) ([]*domain.Deployment, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceRead); err != nil {
		return nil, err
	}

	return s.repo.ListDeployments(ctx, userID)
}

func (s *ContainerService) GetDeployment(ctx context.Context, id uuid.UUID) (*domain.Deployment, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceRead); err != nil {
		return nil, err
	}

	return s.repo.GetDeploymentByID(ctx, id, userID)
}

func (s *ContainerService) ScaleDeployment(ctx context.Context, id uuid.UUID, replicas int) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceUpdate); err != nil {
		return err
	}

	dep, err := s.repo.GetDeploymentByID(ctx, id, userID)
	if err != nil {
		return err
	}

	dep.Replicas = replicas
	dep.Status = domain.DeploymentStatusScaling
	if err := s.repo.UpdateDeployment(ctx, dep); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, userID, "container.deployment_scale", "deployment", id.String(), map[string]interface{}{
		"replicas": replicas,
	})

	return nil
}

func (s *ContainerService) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	if err := s.rbacSvc.Authorize(ctx, userID, tenantID, domain.PermissionInstanceTerminate); err != nil {
		return err
	}

	dep, err := s.repo.GetDeploymentByID(ctx, id, userID)
	if err != nil {
		return err
	}

	dep.Status = domain.DeploymentStatusDeleting
	if err := s.repo.UpdateDeployment(ctx, dep); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, userID, "container.deployment_delete", "deployment", id.String(), map[string]interface{}{})

	return nil
}
