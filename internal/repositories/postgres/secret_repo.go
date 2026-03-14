// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"fmt"
	"time"

	stdlib_errors "errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// SecretRepository provides PostgreSQL-backed secret persistence.
type SecretRepository struct {
	db DB
}

// NewSecretRepository creates a SecretRepository using the provided DB.
func NewSecretRepository(db DB) *SecretRepository {
	return &SecretRepository{db: db}
}

func (r *SecretRepository) Create(ctx context.Context, s *domain.Secret) error {
	query := `
		INSERT INTO secrets (id, user_id, tenant_id, name, encrypted_value, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.UserID, s.TenantID, s.Name, s.EncryptedValue, s.Description, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}
	return nil
}

func (r *SecretRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Secret, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at
		FROM secrets
		WHERE id = $1 AND (tenant_id = $2 OR tenant_id IS NULL)
	`
	return r.scanSecret(r.db.QueryRow(ctx, query, id, tenantID))
}

func (r *SecretRepository) GetByName(ctx context.Context, name string) (*domain.Secret, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at
		FROM secrets
		WHERE name = $1 AND (tenant_id = $2 OR tenant_id IS NULL)
	`
	return r.scanSecret(r.db.QueryRow(ctx, query, name, tenantID))
}

func (r *SecretRepository) List(ctx context.Context) ([]*domain.Secret, error) {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `
		SELECT id, user_id, tenant_id, name, encrypted_value, description, created_at, updated_at, last_accessed_at
		FROM secrets
		WHERE tenant_id = $1 OR tenant_id IS NULL
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	return r.scanSecrets(rows)
}

func (r *SecretRepository) scanSecret(row pgx.Row) (*domain.Secret, error) {
	s := &domain.Secret{}
	err := row.Scan(
		&s.ID, &s.UserID, &s.TenantID, &s.Name, &s.EncryptedValue, &s.Description, &s.CreatedAt, &s.UpdatedAt, &s.LastAccessedAt,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, "secret not found")
		}
		return nil, fmt.Errorf("failed to scan secret: %w", err)
	}
	return s, nil
}

func (r *SecretRepository) scanSecrets(rows pgx.Rows) ([]*domain.Secret, error) {
	defer rows.Close()
	var secrets []*domain.Secret
	for rows.Next() {
		s, err := r.scanSecret(rows)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
	}
	return secrets, nil
}

func (r *SecretRepository) Update(ctx context.Context, s *domain.Secret) error {
	query := `
		UPDATE secrets
		SET encrypted_value = $1, description = $2, updated_at = $3, last_accessed_at = $4
		WHERE id = $5 AND (tenant_id = $6 OR (tenant_id IS NULL AND $6 IS NULL))
	`
	var tenantParam interface{} = s.TenantID
	if s.TenantID == uuid.Nil {
		tenantParam = nil
	}

	_, err := r.db.Exec(ctx, query, s.EncryptedValue, s.Description, time.Now(), s.LastAccessedAt, s.ID, tenantParam)
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}
	return nil
}

func (r *SecretRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID := appcontext.TenantIDFromContext(ctx)
	query := `DELETE FROM secrets WHERE id = $1 AND (tenant_id = $2 OR (tenant_id IS NULL AND $2 IS NULL))`

	var tenantParam interface{} = tenantID
	if tenantID == uuid.Nil {
		tenantParam = nil
	}

	_, err := r.db.Exec(ctx, query, id, tenantParam)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}
