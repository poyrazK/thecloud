package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// TenantRepository is a mock for ports.TenantRepository
type TenantRepository struct {
	mock.Mock
}

func (m *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

func (m *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TenantRepository) AddMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	args := m.Called(ctx, tenantID, userID, role)
	return args.Error(0)
}

func (m *TenantRepository) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	args := m.Called(ctx, tenantID, userID)
	return args.Error(0)
}

func (m *TenantRepository) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]domain.TenantMember, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]domain.TenantMember), args.Error(1)
}

func (m *TenantRepository) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantMember), args.Error(1)
}

func (m *TenantRepository) ListUserTenants(ctx context.Context, userID uuid.UUID) ([]domain.Tenant, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Tenant), args.Error(1)
}

func (m *TenantRepository) GetQuota(ctx context.Context, tenantID uuid.UUID) (*domain.TenantQuota, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantQuota), args.Error(1)
}

func (m *TenantRepository) UpdateQuota(ctx context.Context, quota *domain.TenantQuota) error {
	args := m.Called(ctx, quota)
	return args.Error(0)
}

func (m *TenantRepository) IncrementUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error {
	args := m.Called(ctx, tenantID, resource, amount)
	return args.Error(0)
}

func (m *TenantRepository) DecrementUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error {
	args := m.Called(ctx, tenantID, resource, amount)
	return args.Error(0)
}
