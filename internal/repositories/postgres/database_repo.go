// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
)

// DatabaseRepository provides PostgreSQL-backed database persistence.
type DatabaseRepository struct {
	db DB
}

// NewDatabaseRepository creates a DatabaseRepository using the provided DB.
func NewDatabaseRepository(db DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

func (r *DatabaseRepository) Create(ctx context.Context, db *domain.Database) error {
	query := `
		INSERT INTO databases (id, user_id, name, engine, version, status, vpc_id, container_id, port, username, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		db.ID, db.UserID, db.Name, db.Engine, db.Version, db.Status, db.VpcID, db.ContainerID, db.Port, db.Username, db.Password, db.CreatedAt, db.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to create database", err)
	}
	return nil
}

func (r *DatabaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, name, engine, version, status, vpc_id, COALESCE(container_id, ''), port, username, password, created_at, updated_at
		FROM databases
		WHERE id = $1 AND user_id = $2
	`
	return r.scanDatabase(r.db.QueryRow(ctx, query, id, userID))
}

func (r *DatabaseRepository) List(ctx context.Context) ([]*domain.Database, error) {
	userID := appcontext.UserIDFromContext(ctx)
	query := `
		SELECT id, user_id, name, engine, version, status, vpc_id, COALESCE(container_id, ''), port, username, password, created_at, updated_at
		FROM databases
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list databases", err)
	}
	return r.scanDatabases(rows)
}

func (r *DatabaseRepository) scanDatabase(row pgx.Row) (*domain.Database, error) {
	var db domain.Database
	var engine, status string
	err := row.Scan(
		&db.ID, &db.UserID, &db.Name, &engine, &db.Version, &status, &db.VpcID, &db.ContainerID, &db.Port, &db.Username, &db.Password, &db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.NotFound, "database not found")
		}
		return nil, errors.Wrap(errors.Internal, "failed to scan database", err)
	}
	db.Engine = domain.DatabaseEngine(engine)
	db.Status = domain.DatabaseStatus(status)
	return &db, nil
}

func (r *DatabaseRepository) scanDatabases(rows pgx.Rows) ([]*domain.Database, error) {
	defer rows.Close()
	var databases []*domain.Database
	for rows.Next() {
		db, err := r.scanDatabase(rows)
		if err != nil {
			return nil, err
		}
		databases = append(databases, db)
	}
	return databases, nil
}

func (r *DatabaseRepository) Update(ctx context.Context, db *domain.Database) error {
	query := `
		UPDATE databases
		SET name = $1, status = $2, container_id = $3, port = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7
	`
	now := time.Now()
	cmd, err := r.db.Exec(ctx, query, db.Name, db.Status, db.ContainerID, db.Port, now, db.ID, db.UserID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to update database", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, "database not found")
	}
	db.UpdatedAt = now
	return nil
}

func (r *DatabaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	userID := appcontext.UserIDFromContext(ctx)
	query := `DELETE FROM databases WHERE id = $1 AND user_id = $2`
	cmd, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(errors.Internal, "failed to delete database", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New(errors.NotFound, fmt.Sprintf("database %s not found", id))
	}
	return nil
}
