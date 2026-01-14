// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// AuthService handles user registration, authentication, and session management.
type AuthService interface {
	// Register creates a new user account in the system.
	Register(ctx context.Context, email, password, name string) (*domain.User, error)
	// Login validates credentials and returns the user along with an initial programmatic API key.
	Login(ctx context.Context, email, password string) (*domain.User, string, error)
	// ValidateUser ensures a user exists and is authorized to perform actions.
	ValidateUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
}

// UserRepository handles the persistence and retrieval of User entities.
type UserRepository interface {
	// Create persists a new User.
	Create(ctx context.Context, user *domain.User) error
	// GetByEmail retrieves a User by their unique email address.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	// GetByID retrieves a User by their unique UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// Update modifies an existing User's information.
	Update(ctx context.Context, user *domain.User) error
	// List returns all registered users (typically for administrative use).
	List(ctx context.Context) ([]*domain.User, error)
}
