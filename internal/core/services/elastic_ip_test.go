package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupElasticIPServiceTest(t *testing.T) (ports.ElasticIPService, ports.ElasticIPRepository, ports.InstanceRepository, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	eipRepo := postgres.NewElasticIPRepository(db)
	instRepo := postgres.NewInstanceRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewElasticIPService(services.ElasticIPServiceParams{
		Repo:         eipRepo,
		InstanceRepo: instRepo,
		AuditSvc:     auditSvc,
		Logger:       logger,
	})
	return svc, eipRepo, instRepo, ctx
}

func TestElasticIPAllocateSuccess(t *testing.T) {
	svc, repo, _, ctx := setupElasticIPServiceTest(t)

	eip, err := svc.AllocateIP(ctx)
	require.NoError(t, err)
	require.NotNil(t, eip)
	assert.Equal(t, domain.EIPStatusAllocated, eip.Status)
	assert.NotEmpty(t, eip.PublicIP)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, eip.ID)
	assert.NoError(t, err)
	assert.Equal(t, eip.ID, fetched.ID)
}

func TestElasticIPAssociateSuccess(t *testing.T) {
	svc, _, instRepo, ctx := setupElasticIPServiceTest(t)

	// 1. Setup instance
	inst := &domain.Instance{
		ID:        uuid.New(),
		UserID:    appcontext.UserIDFromContext(ctx),
		TenantID:  appcontext.TenantIDFromContext(ctx),
		Name:      "test-instance",
		Status:    domain.StatusRunning,
		CreatedAt: time.Now(),
	}
	err := instRepo.Create(ctx, inst)
	require.NoError(t, err)

	// 2. Allocate EIP
	eip, err := svc.AllocateIP(ctx)
	require.NoError(t, err)

	// 3. Associate
	associated, err := svc.AssociateIP(ctx, eip.ID, inst.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.EIPStatusAssociated, associated.Status)
	assert.Equal(t, &inst.ID, associated.InstanceID)

	// 4. Disassociate
	disassociated, err := svc.DisassociateIP(ctx, eip.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.EIPStatusAllocated, disassociated.Status)
	assert.Nil(t, disassociated.InstanceID)
}

func TestElasticIPReleaseFailureAssociated(t *testing.T) {
	svc, _, instRepo, ctx := setupElasticIPServiceTest(t)

	inst := &domain.Instance{
		ID:        uuid.New(),
		UserID:    appcontext.UserIDFromContext(ctx),
		TenantID:  appcontext.TenantIDFromContext(ctx),
		Name:      "test",
		Status:    domain.StatusRunning,
		CreatedAt: time.Now(),
	}
	err := instRepo.Create(ctx, inst)
	require.NoError(t, err)

	eip, err := svc.AllocateIP(ctx)
	require.NoError(t, err)

	_, err = svc.AssociateIP(ctx, eip.ID, inst.ID)
	require.NoError(t, err)

	// Should fail release because associated
	err = svc.ReleaseIP(ctx, eip.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disassociate it first")
}

func TestElasticIPReleaseSuccess(t *testing.T) {
	svc, repo, _, ctx := setupElasticIPServiceTest(t)
	eip, err := svc.AllocateIP(ctx)
	require.NoError(t, err)

	err = svc.ReleaseIP(ctx, eip.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, eip.ID)
	assert.Error(t, err)
}

func TestElasticIPListSuccess(t *testing.T) {
	svc, _, _, ctx := setupElasticIPServiceTest(t)
	_, err := svc.AllocateIP(ctx)
	require.NoError(t, err)
	_, err = svc.AllocateIP(ctx)
	require.NoError(t, err)

	eips, err := svc.ListElasticIPs(ctx)
	assert.NoError(t, err)
	assert.Len(t, eips, 2)
}

func TestElasticIPGetSuccess(t *testing.T) {
	svc, _, _, ctx := setupElasticIPServiceTest(t)
	eip, err := svc.AllocateIP(ctx)
	require.NoError(t, err)

	fetched, err := svc.GetElasticIP(ctx, eip.ID)
	assert.NoError(t, err)
	assert.Equal(t, eip.ID, fetched.ID)
}
