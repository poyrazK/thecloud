// Package ports defines interfaces for adapters and services.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// TenantRepository defines the operations for tenant data persistence.
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error
	ListMembers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantMember, error)
	GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error)
	ListUserTenants(ctx context.Context, userID uuid.UUID) ([]domain.Tenant, error)

	GetQuota(ctx context.Context, tenantID uuid.UUID) (*domain.TenantQuota, error)
	UpdateQuota(ctx context.Context, quota *domain.TenantQuota) error
}

// TenantService defines the business logic for tenant management.
type TenantService interface {
	CreateTenant(ctx context.Context, name, slug string, ownerID uuid.UUID) (*domain.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	InviteMember(ctx context.Context, tenantID uuid.UUID, email, role string) error
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error
	SwitchTenant(ctx context.Context, userID, tenantID uuid.UUID) error
	CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, requested int) error
	GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error)
}
