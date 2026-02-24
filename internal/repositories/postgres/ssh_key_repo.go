package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	stdlib_errors "errors"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

type SSHKeyRepo struct {
	db DB
}

func NewSSHKeyRepo(db DB) *SSHKeyRepo {
	return &SSHKeyRepo{db: db}
}

func (r *SSHKeyRepo) Create(ctx context.Context, key *domain.SSHKey) error {
	query := `
		INSERT INTO ssh_keys (id, user_id, tenant_id, name, public_key, fingerprint, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		key.ID,
		key.UserID,
		key.TenantID,
		key.Name,
		key.PublicKey,
		key.Fingerprint,
		key.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create ssh key: %w", err)
	}
	return nil
}

func (r *SSHKeyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SSHKey, error) {
	query := `
		SELECT id, user_id, tenant_id, name, public_key, fingerprint, created_at
		FROM ssh_keys
		WHERE id = $1
	`
	return r.scanSSHKey(r.db.QueryRow(ctx, query, id))
}

func (r *SSHKeyRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.SSHKey, error) {
	query := `
		SELECT id, user_id, tenant_id, name, public_key, fingerprint, created_at
		FROM ssh_keys
		WHERE tenant_id = $1 AND name = $2
	`
	return r.scanSSHKey(r.db.QueryRow(ctx, query, tenantID, name))
}

func (r *SSHKeyRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.SSHKey, error) {
	query := `
		SELECT id, user_id, tenant_id, name, public_key, fingerprint, created_at
		FROM ssh_keys
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list ssh keys: %w", err)
	}
	defer rows.Close()

	var keys []*domain.SSHKey
	for rows.Next() {
		key, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ssh keys: %w", err)
	}

	return keys, nil
}

func (r *SSHKeyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ssh_keys WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete ssh key: %w", err)
	}
	return nil
}

func (r *SSHKeyRepo) scanSSHKey(row pgx.Row) (*domain.SSHKey, error) {
	key := &domain.SSHKey{}
	err := row.Scan(
		&key.ID,
		&key.UserID,
		&key.TenantID,
		&key.Name,
		&key.PublicKey,
		&key.Fingerprint,
		&key.CreatedAt,
	)
	if err != nil {
		if stdlib_errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.NotFound, "ssh key not found")
		}
		return nil, fmt.Errorf("failed to scan ssh key: %w", err)
	}
	return key, nil
}

func (r *SSHKeyRepo) scanRow(rows pgx.Rows) (*domain.SSHKey, error) {
	key := &domain.SSHKey{}
	err := rows.Scan(
		&key.ID,
		&key.UserID,
		&key.TenantID,
		&key.Name,
		&key.PublicKey,
		&key.Fingerprint,
		&key.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}
	return key, nil
}
