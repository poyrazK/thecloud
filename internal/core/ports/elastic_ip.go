// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// ElasticIPRepository handles the persistence and retrieval of Elastic IP metadata.
type ElasticIPRepository interface {
	// Create saves a new Elastic IP record.
	Create(ctx context.Context, eip *domain.ElasticIP) error
	// GetByID retrieves an Elastic IP by its unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error)
	// GetByPublicIP retrieves an Elastic IP by its public address.
	GetByPublicIP(ctx context.Context, publicIP string) (*domain.ElasticIP, error)
	// GetByInstanceID retrieves the Elastic IP associated with a specific instance.
	GetByInstanceID(ctx context.Context, instanceID uuid.UUID) (*domain.ElasticIP, error)
	// List returns Elastic IPs authorized for the current operational context.
	List(ctx context.Context) ([]*domain.ElasticIP, error)
	// Update modifies an existing Elastic IP's metadata or status.
	Update(ctx context.Context, eip *domain.ElasticIP) error
	// Delete removes an Elastic IP record from persistent storage.
	Delete(ctx context.Context, id uuid.UUID) error
}

// ElasticIPService defines the business logic for managing the lifecycle of Elastic IPs.
type ElasticIPService interface {
	// AllocateIP reserves a new Elastic IP for the current user.
	AllocateIP(ctx context.Context) (*domain.ElasticIP, error)
	// ReleaseIP returns an Elastic IP back to the pool.
	ReleaseIP(ctx context.Context, id uuid.UUID) error
	// AssociateIP maps an Elastic IP to a specific compute instance.
	AssociateIP(ctx context.Context, id uuid.UUID, instanceID uuid.UUID) (*domain.ElasticIP, error)
	// DisassociateIP removes the mapping between an Elastic IP and an instance.
	DisassociateIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error)
	// ListElasticIPs returns a slice of all Elastic IPs accessible to the caller.
	ListElasticIPs(ctx context.Context) ([]*domain.ElasticIP, error)
	// GetElasticIP retrieves detailed information about a specific Elastic IP.
	GetElasticIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error)
}
