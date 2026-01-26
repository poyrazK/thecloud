package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testTenantName = "My Tenant"
	testTenantSlug = "my-tenant"
	testSlugOk     = "slug-ok"
)

func setupTenantServiceTest(_ *testing.T) (*MockTenantRepo, *MockUserRepo, *MockAuditService, *services.TenantService) {
	tenantRepo := new(MockTenantRepo)
	userRepo := new(MockUserRepo)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := services.NewTenantService(tenantRepo, userRepo, logger)
	return tenantRepo, userRepo, auditSvc, svc
}

func TestCreateTenantSuccess(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()
	name := testTenantName
	slug := testTenantSlug

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

func TestCreateTenantDoesNotUpdateUserWhenDefaultTenantSet(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()
	name := testTenantName
	slug := testTenantSlug
	defaultTenantID := uuid.New()

	tenantRepo.On("GetBySlug", mock.Anything, slug).Return(nil, nil)
	tenantRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	tenantRepo.On("AddMember", mock.Anything, mock.Anything, ownerID, "owner").Return(nil)
	tenantRepo.On("UpdateQuota", mock.Anything, mock.Anything).Return(nil).Maybe()

	userRepo.On("GetByID", mock.Anything, ownerID).Return(&domain.User{ID: ownerID, DefaultTenantID: &defaultTenantID}, nil)

	tenant, err := svc.CreateTenant(ctx, name, slug, ownerID)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestCreateTenantSlugTaken(t *testing.T) {
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

func TestCreateTenantAddMemberError(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()
	name := testTenantName
	slug := testTenantSlug

	tenantRepo.On("GetBySlug", mock.Anything, slug).Return(nil, nil)
	tenantRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	tenantRepo.On("AddMember", mock.Anything, mock.Anything, ownerID, "owner").Return(assert.AnError)
	userRepo.On("GetByID", mock.Anything, ownerID).Return(&domain.User{ID: ownerID}, nil).Maybe()

	tenant, err := svc.CreateTenant(ctx, name, slug, ownerID)
	assert.Error(t, err)
	assert.Nil(t, tenant)
}

func TestCreateTenantCreateError(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()

	tenantRepo.On("GetBySlug", mock.Anything, "slug-fail").Return(nil, nil)
	tenantRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	tenant, err := svc.CreateTenant(ctx, testTenantName, "slug-fail", ownerID)
	assert.Error(t, err)
	assert.Nil(t, tenant)
}

func TestCreateTenantUpdateQuotaErrorContinues(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()

	tenantRepo.On("GetBySlug", mock.Anything, testSlugOk).Return(nil, nil)
	tenantRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	tenantRepo.On("AddMember", mock.Anything, mock.Anything, ownerID, "owner").Return(nil)
	tenantRepo.On("UpdateQuota", mock.Anything, mock.Anything).Return(assert.AnError).Once()
	userRepo.On("GetByID", mock.Anything, ownerID).Return(&domain.User{ID: ownerID}, nil)
	userRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	tenant, err := svc.CreateTenant(ctx, testTenantName, testSlugOk, ownerID)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
}

func TestCreateTenantUserLookupErrorContinues(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()

	tenantRepo.On("GetBySlug", mock.Anything, testSlugOk).Return(nil, nil)
	tenantRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	tenantRepo.On("AddMember", mock.Anything, mock.Anything, ownerID, "owner").Return(nil)
	tenantRepo.On("UpdateQuota", mock.Anything, mock.Anything).Return(nil).Maybe()
	userRepo.On("GetByID", mock.Anything, ownerID).Return(nil, assert.AnError)

	tenant, err := svc.CreateTenant(ctx, testTenantName, testSlugOk, ownerID)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
}

func TestCreateTenantUserUpdateErrorContinues(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	ownerID := uuid.New()

	tenantRepo.On("GetBySlug", mock.Anything, "slug-ok-2").Return(nil, nil)
	tenantRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	tenantRepo.On("AddMember", mock.Anything, mock.Anything, ownerID, "owner").Return(nil)
	tenantRepo.On("UpdateQuota", mock.Anything, mock.Anything).Return(nil).Maybe()
	userRepo.On("GetByID", mock.Anything, ownerID).Return(&domain.User{ID: ownerID}, nil)
	userRepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError)

	tenant, err := svc.CreateTenant(ctx, testTenantName, "slug-ok-2", ownerID)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
}

func TestInviteMemberSuccess(t *testing.T) {
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

func TestInviteMemberUserNotFound(t *testing.T) {
	_, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	tenantID := uuid.New()
	userEmail := "missing@example.com"

	userRepo.On("GetByEmail", mock.Anything, userEmail).Return(nil, assert.AnError)

	err := svc.InviteMember(ctx, tenantID, userEmail, "member")
	assert.Error(t, err)
}

func TestInviteMemberAlreadyMember(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	tenantID := uuid.New()
	userEmail := "invitee@example.com"
	userID := uuid.New()

	userRepo.On("GetByEmail", mock.Anything, userEmail).Return(&domain.User{ID: userID, Email: userEmail}, nil)
	tenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{UserID: userID, TenantID: tenantID}, nil)

	err := svc.InviteMember(ctx, tenantID, userEmail, "member")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already a member")
}

func TestSwitchTenantSuccess(t *testing.T) {
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

func TestSwitchTenantNotMember(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	tenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(nil, assert.AnError)

	err := svc.SwitchTenant(ctx, userID, tenantID)
	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func TestSwitchTenantGetUserError(t *testing.T) {
	tenantRepo, userRepo, _, svc := setupTenantServiceTest(t)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	tenantRepo.On("GetMembership", mock.Anything, tenantID, userID).Return(&domain.TenantMember{UserID: userID, TenantID: tenantID}, nil)
	userRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	err := svc.SwitchTenant(ctx, userID, tenantID)
	assert.Error(t, err)
}

func TestCheckQuotaWithinLimit(t *testing.T) {
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

func TestCheckQuotaExceeded(t *testing.T) {
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

func TestGetTenantSuccess(t *testing.T) {
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

func TestRemoveMemberSuccess(t *testing.T) {
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

func TestRemoveMemberCannotRemoveOwner(t *testing.T) {
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

func TestGetMembershipSuccess(t *testing.T) {
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

func TestCheckQuotaResources(t *testing.T) {
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

func TestCheckQuotaResourcesWithinLimit(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()

	quota := &domain.TenantQuota{
		TenantID:      tenantID,
		MaxStorageGB:  50,
		UsedStorageGB: 10,
		MaxMemoryGB:   16,
		UsedMemoryGB:  8,
		MaxVCPUs:      8,
		UsedVCPUs:     4,
		MaxVPCs:       2,
		UsedVPCs:      1,
		MaxInstances:  10,
		UsedInstances: 3,
	}
	tenantRepo.On("GetQuota", mock.Anything, tenantID).Return(quota, nil)

	assert.NoError(t, svc.CheckQuota(ctx, tenantID, "storage", 5))
	assert.NoError(t, svc.CheckQuota(ctx, tenantID, "memory", 4))
	assert.NoError(t, svc.CheckQuota(ctx, tenantID, "vcpus", 2))
}

func TestRemoveMemberRepoError(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()

	tenant := &domain.Tenant{ID: tenantID, OwnerID: ownerID}
	tenantRepo.On("GetByID", mock.Anything, tenantID).Return(tenant, nil)
	tenantRepo.On("RemoveMember", mock.Anything, tenantID, userID).Return(assert.AnError)

	err := svc.RemoveMember(ctx, tenantID, userID)
	assert.Error(t, err)
}

func TestRemoveMemberGetTenantError(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()

	tenantRepo.On("GetByID", mock.Anything, tenantID).Return(nil, assert.AnError)

	err := svc.RemoveMember(ctx, tenantID, userID)
	assert.Error(t, err)
}

func TestCheckQuotaGetQuotaError(t *testing.T) {
	tenantRepo, _, _, svc := setupTenantServiceTest(t)
	ctx := context.Background()
	tenantID := uuid.New()

	tenantRepo.On("GetQuota", mock.Anything, tenantID).Return(nil, assert.AnError)

	err := svc.CheckQuota(ctx, tenantID, "instances", 1)
	assert.Error(t, err)
}
