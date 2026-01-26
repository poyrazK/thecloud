package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTenantServiceTest(_ *testing.T) (*MockTenantRepo, *MockUserRepo, *MockAuditService, *services.TenantService) {
	tenantRepo := new(MockTenantRepo)
	userRepo := new(MockUserRepo)
	auditSvc := new(MockAuditService)
	// We pass nil for logger for now, or you can mock it if needed
	svc := services.NewTenantService(tenantRepo, userRepo, nil)
	return tenantRepo, userRepo, auditSvc, svc
}

func TestCreateTenant_Success(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()
	name := "My Tenant"
	slug := "my-tenant"

	tenantRepo.On("GetBySlug", mock.Anything, slug).Return(nil, nil) // Slug available
	tenantRepo.On("Create", mock.Anything, mock.MatchedBy(func(tenant *domain.Tenant) bool {
		return tenant.Name == name && tenant.Slug == slug && tenant.OwnerID == ownerID
	})).Return(nil)
	tenantRepo.On("AddMember", mock.Anything, mock.Anything, ownerID, "owner").Return(nil)
	tenantRepo.On("UpdateQuota", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Mock user update for default tenant
	userRepo.On("GetByID", mock.Anything, ownerID).Return(&domain.User{ID: ownerID, DefaultTenantID: nil}, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == ownerID // We could verify DefaultTenantID is set too
	})).Return(nil)

	tenant, err := svc.CreateTenant(ctx, name, slug, ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, name, tenant.Name)
	assert.Equal(t, slug, tenant.Slug)
}

func TestCreateTenant_SlugTaken(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()
	slug := "taken-slug"

	existingTenant := &domain.Tenant{ID: uuid.New(), Slug: slug}
	tenantRepo.On("GetBySlug", mock.Anything, slug).Return(existingTenant, nil)

	tenant, err := svc.CreateTenant(ctx, "Name", slug, ownerID)

	assert.Error(t, err)
	assert.Nil(t, tenant)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInviteMember_Success(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	tenantID := uuid.New()
	userEmail := "invitee@example.com"
	userID := uuid.New()

	existingUser := &domain.User{ID: userID, Email: userEmail}

	// Mock finding user
	userRepo.On("GetByEmail", mock.Anything, userEmail).Return(existingUser, nil)
	// Mock checking if already member (returns nil if not found, or error if not found?
	// Usually repo returns proper error for not found, let's assume service handles "not found" gracefully or we return specific error)
	// Actually typical implementation checks if member exists. Let's assume GetMembership returns error if not found.
	// But mocking success flow:
	tenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(nil, assert.AnError)

	tenantRepo.On("AddMember", mock.Anything, tenantID, userID, "member").Return(nil)

	err := svc.InviteMember(ctx, tenantID, userEmail, "member")
	assert.NoError(t, err)
}

func TestSwitchTenant_Success(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	// Tenant exists
	tenantRepo.On("GetByID", mock.Anything, tenantID).Return(&domain.Tenant{ID: tenantID}, nil)
	// User is member
	tenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{UserID: userID, TenantID: tenantID}, nil)

	// User update default tenant
	userRepo.On("GetByID", mock.Anything, userID).Return(&domain.User{ID: userID}, nil)
	userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return *u.DefaultTenantID == tenantID
	})).Return(nil)

	err := svc.SwitchTenant(ctx, userID, tenantID)
	assert.NoError(t, err)
}

func TestCheckQuota_WithinLimit(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()

	quota := &domain.TenantQuota{
		TenantID:      tenantID,
		MaxInstances:  10,
		UsedInstances: 5,
	}
	tenantRepo.On("GetQuota", mock.Anything, tenantID).Return(quota, nil)

	err := svc.CheckQuota(ctx, tenantID, "instances", 1)
	assert.NoError(t, err)
}

func TestCheckQuota_Exceeded(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()

	quota := &domain.TenantQuota{
		TenantID:      tenantID,
		MaxInstances:  10,
		UsedInstances: 10,
	}
	tenantRepo.On("GetQuota", mock.Anything, tenantID).Return(quota, nil)

	err := svc.CheckQuota(ctx, tenantID, "instances", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quota exceeded")
}

func TestGetTenant_Success(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()
	tenant := &domain.Tenant{ID: tenantID, Name: "Found Tenant"}

	tenantRepo.On("GetByID", mock.Anything, tenantID).Return(tenant, nil)

	result, err := svc.GetTenant(ctx, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, tenantID, result.ID)
	assert.Equal(t, "Found Tenant", result.Name)
}

func TestRemoveMember_Success(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()

	// Mock getting tenant to check owner
	tenant := &domain.Tenant{ID: tenantID, OwnerID: ownerID} // User is NOT owner
	tenantRepo.On("GetByID", mock.Anything, tenantID).Return(tenant, nil)

	tenantRepo.On("RemoveMember", mock.Anything, tenantID, userID).Return(nil)

	err := svc.RemoveMember(ctx, tenantID, userID)
	assert.NoError(t, err)
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()
	ownerID := uuid.New()

	// Mock getting tenant to check owner
	tenant := &domain.Tenant{ID: tenantID, OwnerID: ownerID}
	tenantRepo.On("GetByID", mock.Anything, tenantID).Return(tenant, nil)

	err := svc.RemoveMember(ctx, tenantID, ownerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove tenant owner")
}

func TestGetMembership_Success(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	member := &domain.TenantMember{TenantID: tenantID, UserID: userID, Role: "member"}

	tenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(member, nil)

	result, err := svc.GetMembership(ctx, tenantID, userID)
	assert.NoError(t, err)
	assert.Equal(t, member, result)
}

func TestCheckQuota_Resources(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()

	quota := &domain.TenantQuota{
		TenantID:      tenantID,
		MaxInstances:  10,
		UsedInstances: 5,
		MaxVPCs:       2,
		UsedVPCs:      2,
	}
	tenantRepo.On("GetQuota", mock.Anything, tenantID).Return(quota, nil)

	// Instances OK
	assert.NoError(t, svc.CheckQuota(ctx, tenantID, "instances", 1))
	// VPCs Full
	err := svc.CheckQuota(ctx, tenantID, "vpcs", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quota exceeded")
	// Unknown resource
	err = svc.CheckQuota(ctx, tenantID, "unknown", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown resource type")
}
