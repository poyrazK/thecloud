package setup

import (
	"log/slog"
	"os"
	"testing"

	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/stretchr/testify/assert"
)

func TestInitIdentityServices(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := &platform.Config{}
	repos := &Repositories{}

	// This will likely need mocks in Repositories
	// But let's see if we can test the branching logic
	svc := initIdentityServices(ServiceConfig{
		Config: cfg,
		Repos:  repos,
		Logger: logger,
	}, nil)

	assert.NotNil(t, svc)
}

func TestInitRBACServices(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := &platform.Config{}
	repos := &Repositories{}

	svc := initRBACServices(ServiceConfig{
		Config: cfg,
		Repos:  repos,
		Logger: logger,
	})

	assert.NotNil(t, svc)
}
