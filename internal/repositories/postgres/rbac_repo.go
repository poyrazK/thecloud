// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type rbacRepository struct {
	db DB
}

// NewRBACRepository creates an RBAC repository using the provided DB.
func NewRBACRepository(db DB) ports.RoleRepository {
	return &rbacRepository{db: db}
}

func (r *rbacRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, "INSERT INTO roles (id, name, description) VALUES ($1, $2, $3)",
		role.ID, role.Name, role.Description)
	if err != nil {
		return err
	}

	for _, p := range role.Permissions {
		_, err = tx.Exec(ctx, "INSERT INTO role_permissions (role_id, permission) VALUES ($1, $2)",
			role.ID, string(p))
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *rbacRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role, err := r.scanRole(r.db.QueryRow(ctx, "SELECT id, name, description FROM roles WHERE id = $1", id))
	if err != nil {
		return nil, err
	}

	perms, err := r.GetPermissionsForRole(ctx, id)
	if err != nil {
		return nil, err
	}
	role.Permissions = perms
	return role, nil
}

func (r *rbacRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	role, err := r.scanRole(r.db.QueryRow(ctx, "SELECT id, name, description FROM roles WHERE name = $1", name))
	if err != nil {
		return nil, err
	}

	perms, err := r.GetPermissionsForRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = perms
	return role, nil
}

func (r *rbacRepository) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, description FROM roles")
	if err != nil {
		return nil, err
	}
	roles, err := r.scanRoles(rows)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		perms, err := r.GetPermissionsForRole(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = perms
	}

	return roles, nil
}

func (r *rbacRepository) scanRole(row pgx.Row) (*domain.Role, error) {
	role := &domain.Role{}
	if err := row.Scan(&role.ID, &role.Name, &role.Description); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *rbacRepository) scanRoles(rows pgx.Rows) ([]*domain.Role, error) {
	defer rows.Close()
	var roles []*domain.Role
	for rows.Next() {
		role, err := r.scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *rbacRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, "UPDATE roles SET name = $1, description = $2 WHERE id = $3",
		role.Name, role.Description, role.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "DELETE FROM role_permissions WHERE role_id = $1", role.ID)
	if err != nil {
		return err
	}

	for _, p := range role.Permissions {
		_, err = tx.Exec(ctx, "INSERT INTO role_permissions (role_id, permission) VALUES ($1, $2)",
			role.ID, string(p))
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *rbacRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	// role_permissions should be deleted by cascade or manually
	_, err := r.db.Exec(ctx, "DELETE FROM roles WHERE id = $1", id)
	return err
}

func (r *rbacRepository) AddPermissionToRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	_, err := r.db.Exec(ctx, "INSERT INTO role_permissions (role_id, permission) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		roleID, string(permission))
	return err
}

func (r *rbacRepository) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permission domain.Permission) error {
	_, err := r.db.Exec(ctx, "DELETE FROM role_permissions WHERE role_id = $1 AND permission = $2",
		roleID, string(permission))
	return err
}

func (r *rbacRepository) GetPermissionsForRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	rows, err := r.db.Query(ctx, "SELECT permission FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, domain.Permission(p))
	}
	return perms, nil
}
