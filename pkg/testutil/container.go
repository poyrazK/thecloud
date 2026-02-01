package testutil

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer represents a running PostgreSQL test container
type PostgresContainer struct {
	Container  *postgres.PostgresContainer
	ConnString string
}

// SetupPostgresContainer starts a PostgreSQL container for testing
func SetupPostgresContainer(t *testing.T) (*PostgresContainer, func()) {
	t.Helper()

	ctx := context.Background()

	// Disable default logger to avoid emojis in output
	discardLogger := log.New(io.Discard, "", 0)

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithLogger(discardLogger),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	container := &PostgresContainer{
		Container:  pgContainer,
		ConnString: connStr,
	}

	cleanup := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}

	return container, cleanup
}
