package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestRBACRepository_CreateRole(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewRBACRepository(mock)
	role := &domain.Role{
		ID:          uuid.New(),
		Name:        "test-role",
		Description: "description",
		Permissions: []domain.Permission{domain.PermissionInstanceRead},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO roles").
		WithArgs(role.ID, role.Name, role.Description).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO role_permissions").
		WithArgs(role.ID, string(role.Permissions[0])).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	err = repo.CreateRole(context.Background(), role)
	assert.NoError(t, err)
}

func TestRBACRepository_GetRoleByID(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewRBACRepository(mock)
	id := uuid.New()

	mock.ExpectQuery("SELECT id, name, description FROM roles").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description"}).
			AddRow(id, "test-role", "description"))

	mock.ExpectQuery("SELECT permission FROM role_permissions").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"permission"}).
			AddRow(string(domain.PermissionInstanceRead)))

	role, err := repo.GetRoleByID(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, role)
	assert.Equal(t, id, role.ID)
	assert.Len(t, role.Permissions, 1)
}

func TestRBACRepository_ListRoles(t *testing.T) {
	t.Parallel()
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewRBACRepository(mock)
	id := uuid.New()

	mock.ExpectQuery("SELECT id, name, description FROM roles").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description"}).
			AddRow(id, "test-role", "description"))

	mock.ExpectQuery("SELECT permission FROM role_permissions").
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{"permission"}).
			AddRow(string(domain.PermissionInstanceRead)))

	roles, err := repo.ListRoles(context.Background())
	assert.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Len(t, roles[0].Permissions, 1)
}
