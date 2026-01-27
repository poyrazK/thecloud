package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantRepoCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	ten := &domain.Tenant{
		ID:        uuid.New(),
		Name:      "Tenant",
		Slug:      "tenant",
		OwnerID:   uuid.New(),
		Plan:      "free",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO tenants").
		WithArgs(ten.ID, ten.Name, ten.Slug, ten.OwnerID, ten.Plan, ten.Status, ten.CreatedAt, ten.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.Create(context.Background(), ten)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepoGetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	id := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT id, name, slug, owner_id, plan, status, created_at, updated_at FROM tenants WHERE id = \$1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "slug", "owner_id", "plan", "status", "created_at", "updated_at"}).
			AddRow(id, "Tenant", "tenant", uuid.New(), "free", "active", now, now))

	tenant, err := repo.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, id, tenant.ID)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(`SELECT id, name, slug, owner_id, plan, status, created_at, updated_at FROM tenants WHERE id = \$1`).
		WithArgs(id).
		WillReturnError(pgx.ErrNoRows)

	tenant, err = repo.GetByID(context.Background(), id)
	assert.Error(t, err)
	assert.Nil(t, tenant)
}

func TestTenantRepoGetBySlug(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	slug := "tenant"
	now := time.Now()

	mock.ExpectQuery(`SELECT id, name, slug, owner_id, plan, status, created_at, updated_at FROM tenants WHERE slug = \$1`).
		WithArgs(slug).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "slug", "owner_id", "plan", "status", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Tenant", slug, uuid.New(), "free", "active", now, now))

	tenant, err := repo.GetBySlug(context.Background(), slug)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, slug, tenant.Slug)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(`SELECT id, name, slug, owner_id, plan, status, created_at, updated_at FROM tenants WHERE slug = \$1`).
		WithArgs(slug).
		WillReturnError(pgx.ErrNoRows)

	tenant, err = repo.GetBySlug(context.Background(), slug)
	assert.Error(t, err)
	assert.Nil(t, tenant)
}

func TestTenantRepoUpdateDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	ten := &domain.Tenant{
		ID:        uuid.New(),
		Name:      "New",
		Slug:      "new",
		OwnerID:   uuid.New(),
		Plan:      "pro",
		Status:    "active",
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec("UPDATE tenants").
		WithArgs(ten.ID, ten.Name, ten.Slug, ten.OwnerID, ten.Plan, ten.Status, ten.UpdatedAt).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.Update(context.Background(), ten)
	assert.NoError(t, err)

	mock.ExpectExec(`DELETE FROM tenants WHERE id = \$1`).
		WithArgs(ten.ID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(context.Background(), ten.ID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepoMembers(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	tenantID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec("INSERT INTO tenant_members").
		WithArgs(tenantID, userID, "owner").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.AddMember(context.Background(), tenantID, userID, "owner")
	assert.NoError(t, err)

	mock.ExpectExec(`DELETE FROM tenant_members WHERE tenant_id = \$1 AND user_id = \$2`).
		WithArgs(tenantID, userID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.RemoveMember(context.Background(), tenantID, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepoListMembers(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	tenantID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT tenant_id, user_id, role, joined_at FROM tenant_members WHERE tenant_id = \$1`).
		WithArgs(tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"tenant_id", "user_id", "role", "joined_at"}).
			AddRow(tenantID, uuid.New(), "owner", now))

	members, err := repo.ListMembers(context.Background(), tenantID)
	assert.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, "owner", members[0].Role)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(`SELECT tenant_id, user_id, role, joined_at FROM tenant_members WHERE tenant_id = \$1`).
		WithArgs(tenantID).
		WillReturnRows(pgxmock.NewRows([]string{"tenant_id", "user_id", "role", "joined_at"}).
			AddRow("bad", uuid.New(), "owner", now))

	members, err = repo.ListMembers(context.Background(), tenantID)
	assert.Error(t, err)
	assert.Nil(t, members)
}

func TestTenantRepoGetMembership(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT tenant_id, user_id, role, joined_at FROM tenant_members WHERE tenant_id = \$1 AND user_id = \$2`).
		WithArgs(tenantID, userID).
		WillReturnRows(pgxmock.NewRows([]string{"tenant_id", "user_id", "role", "joined_at"}).
			AddRow(tenantID, userID, "member", now))

	member, err := repo.GetMembership(context.Background(), tenantID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, "member", member.Role)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(`SELECT tenant_id, user_id, role, joined_at FROM tenant_members WHERE tenant_id = \$1 AND user_id = \$2`).
		WithArgs(tenantID, userID).
		WillReturnError(pgx.ErrNoRows)

	member, err = repo.GetMembership(context.Background(), tenantID, userID)
	assert.NoError(t, err)
	assert.Nil(t, member)
}

func TestTenantRepoListUserTenants(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`FROM tenants t\s+JOIN tenant_members tm ON t.id = tm.tenant_id\s+WHERE tm.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "slug", "owner_id", "plan", "status", "created_at", "updated_at"}).
			AddRow(uuid.New(), "Tenant", "tenant", uuid.New(), "free", "active", now, now))

	tenants, err := repo.ListUserTenants(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, tenants, 1)
	assert.Equal(t, "tenant", tenants[0].Slug)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(`FROM tenants t\s+JOIN tenant_members tm ON t.id = tm.tenant_id\s+WHERE tm.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "slug", "owner_id", "plan", "status", "created_at", "updated_at"}).
			AddRow("bad", "Tenant", "tenant", uuid.New(), "free", "active", now, now))

	tenants, err = repo.ListUserTenants(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, tenants)
}

func TestTenantRepoGetQuotaUpdateQuota(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	tenantID := uuid.New()

	mock.ExpectQuery("FROM tenant_quotas tq").
		WithArgs(tenantID).
		WillReturnRows(pgxmock.NewRows([]string{
			"tenant_id",
			"max_instances",
			"max_vpcs",
			"max_storage_gb",
			"max_memory_gb",
			"max_vcpus",
			"used_instances",
			"used_vpcs",
			"used_storage_gb",
			"used_memory_gb",
			"used_vcpus",
		}).AddRow(tenantID, 10, 5, 100, 32, 16, 2, 1, 50, 0, 0))

	quota, err := repo.GetQuota(context.Background(), tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, quota)
	assert.Equal(t, 10, quota.MaxInstances)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery("FROM tenant_quotas tq").
		WithArgs(tenantID).
		WillReturnError(pgx.ErrNoRows)

	quota, err = repo.GetQuota(context.Background(), tenantID)
	assert.Error(t, err)
	assert.Nil(t, quota)

	q := &domain.TenantQuota{
		TenantID:     tenantID,
		MaxInstances: 10,
		MaxVPCs:      5,
		MaxStorageGB: 100,
		MaxMemoryGB:  32,
		MaxVCPUs:     16,
	}

	mock.ExpectExec("INSERT INTO tenant_quotas").
		WithArgs(q.TenantID, q.MaxInstances, q.MaxVPCs, q.MaxStorageGB, q.MaxMemoryGB, q.MaxVCPUs).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.UpdateQuota(context.Background(), q)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTenantRepoExecErrors(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewTenantRepo(mock)
	ten := &domain.Tenant{ID: uuid.New()}

	mock.ExpectExec("INSERT INTO tenants").WillReturnError(errors.New("db"))
	err = repo.Create(context.Background(), ten)
	assert.Error(t, err)

	mock.ExpectExec("UPDATE tenants").WillReturnError(errors.New("db"))
	err = repo.Update(context.Background(), ten)
	assert.Error(t, err)

	mock.ExpectExec(`DELETE FROM tenants WHERE id = \$1`).WithArgs(ten.ID).WillReturnError(errors.New("db"))
	err = repo.Delete(context.Background(), ten.ID)
	assert.Error(t, err)

	mock.ExpectExec("INSERT INTO tenant_members").WillReturnError(errors.New("db"))
	err = repo.AddMember(context.Background(), uuid.New(), uuid.New(), "member")
	assert.Error(t, err)

	mock.ExpectExec(`DELETE FROM tenant_members WHERE tenant_id = \$1 AND user_id = \$2`).WillReturnError(errors.New("db"))
	err = repo.RemoveMember(context.Background(), uuid.New(), uuid.New())
	assert.Error(t, err)

	mock.ExpectQuery(`SELECT tenant_id, user_id, role, joined_at FROM tenant_members WHERE tenant_id = \$1`).
		WithArgs(uuid.New()).
		WillReturnError(errors.New("db"))
	members, err := repo.ListMembers(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, members)

	mock.ExpectQuery(`FROM tenants t\s+JOIN tenant_members tm ON t.id = tm.tenant_id\s+WHERE tm.user_id = \$1`).
		WithArgs(uuid.New()).
		WillReturnError(errors.New("db"))
	members2, err := repo.ListUserTenants(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Nil(t, members2)
}
