// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// DatabaseRepository handles the persistence and retrieval of managed database metadata.
type DatabaseRepository interface {
	// Create saves a new database instance record.
	Create(ctx context.Context, db *domain.Database) error
	// GetByID retrieves a database instance by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Database, error)
	// List returns all database instances (typically filtered by current user context).
	List(ctx context.Context) ([]*domain.Database, error)
	// Update modifies an existing database's metadata or status.
	Update(ctx context.Context, db *domain.Database) error
	// Delete removes a database record from storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// DatabaseService provides business logic for managing relational database instances (DBaaS).
type DatabaseService interface {
	// CreateDatabase provisions a new managed database instance.
	CreateDatabase(ctx context.Context, name, engine, version string, vpcID *uuid.UUID) (*domain.Database, error)
	// GetDatabase retrieves details for a specific database.
	GetDatabase(ctx context.Context, id uuid.UUID) (*domain.Database, error)
	// ListDatabases returns all databases for the authorized user.
	ListDatabases(ctx context.Context) ([]*domain.Database, error)
	// DeleteDatabase terminates and deletes a database instance.
	DeleteDatabase(ctx context.Context, id uuid.UUID) error
	// GetConnectionString constructs and returns the authorized URI for connecting to the database.
	GetConnectionString(ctx context.Context, id uuid.UUID) (string, error)
}
