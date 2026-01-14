// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

// PostgresContainerRepository provides PostgreSQL-backed container persistence.
type PostgresContainerRepository struct {
	db DB
}

// NewPostgresContainerRepository creates a container repository using the provided DB.
func NewPostgresContainerRepository(db DB) ports.ContainerRepository {
	return &PostgresContainerRepository{db: db}
}

func (r *PostgresContainerRepository) CreateDeployment(ctx context.Context, d *domain.Deployment) error {
	query := `
		INSERT INTO deployments (id, user_id, name, image, replicas, current_count, ports, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		d.ID,
		d.UserID,
		d.Name,
		d.Image,
		d.Replicas,
		d.CurrentCount,
		d.Ports,
		d.Status,
		d.CreatedAt,
		d.UpdatedAt,
	)
	return err
}

func (r *PostgresContainerRepository) GetDeploymentByID(ctx context.Context, id, userID uuid.UUID) (*domain.Deployment, error) {
	query := `SELECT id, user_id, name, image, replicas, current_count, ports, status, created_at, updated_at FROM deployments WHERE id = $1 AND user_id = $2`
	return r.scanDeployment(r.db.QueryRow(ctx, query, id, userID))
}

func (r *PostgresContainerRepository) ListDeployments(ctx context.Context, userID uuid.UUID) ([]*domain.Deployment, error) {
	query := `SELECT id, user_id, name, image, replicas, current_count, ports, status, created_at, updated_at FROM deployments WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	return r.scanDeployments(rows)
}

func (r *PostgresContainerRepository) UpdateDeployment(ctx context.Context, d *domain.Deployment) error {
	query := `
		UPDATE deployments 
		SET replicas = $1, current_count = $2, status = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, d.Replicas, d.CurrentCount, d.Status, d.ID)
	return err
}

func (r *PostgresContainerRepository) DeleteDeployment(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM deployments WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresContainerRepository) AddContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error {
	query := `INSERT INTO deployment_containers (deployment_id, instance_id) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, query, deploymentID, instanceID)
	return err
}

func (r *PostgresContainerRepository) RemoveContainer(ctx context.Context, deploymentID, instanceID uuid.UUID) error {
	query := `DELETE FROM deployment_containers WHERE deployment_id = $1 AND instance_id = $2`
	_, err := r.db.Exec(ctx, query, deploymentID, instanceID)
	return err
}

func (r *PostgresContainerRepository) GetContainers(ctx context.Context, deploymentID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT instance_id FROM deployment_containers WHERE deployment_id = $1`
	rows, err := r.db.Query(ctx, query, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PostgresContainerRepository) ListAllDeployments(ctx context.Context) ([]*domain.Deployment, error) {
	query := `SELECT id, user_id, name, image, replicas, current_count, ports, status, created_at, updated_at FROM deployments`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return r.scanDeployments(rows)
}

func (r *PostgresContainerRepository) scanDeployment(row pgx.Row) (*domain.Deployment, error) {
	var d domain.Deployment
	var status string
	err := row.Scan(
		&d.ID,
		&d.UserID,
		&d.Name,
		&d.Image,
		&d.Replicas,
		&d.CurrentCount,
		&d.Ports,
		&status,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	d.Status = domain.DeploymentStatus(status)
	return &d, nil
}

func (r *PostgresContainerRepository) scanDeployments(rows pgx.Rows) ([]*domain.Deployment, error) {
	defer rows.Close()
	var deps []*domain.Deployment
	for rows.Next() {
		d, err := r.scanDeployment(rows)
		if err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, nil
}
