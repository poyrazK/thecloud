package platform

import (
	"os"
	"testing"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigDefaults(t *testing.T) {
	// Ensure env vars are unset to test defaults
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("APP_ENV")

	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.Equal(t, testutil.TestPort, cfg.Port)
	assert.Equal(t, testutil.TestDatabaseURL, cfg.DatabaseURL)
	assert.Equal(t, testutil.TestEnvDev, cfg.Environment)
}

func TestNewConfigEnvVars(t *testing.T) {
	_ = os.Setenv("PORT", "9090")
	_ = os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
	_ = os.Setenv("APP_ENV", "production")
	defer func() {
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("DATABASE_URL")
		_ = os.Unsetenv("APP_ENV")
	}()

	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.Equal(t, testutil.TestProdPort, cfg.Port)
	assert.Equal(t, testutil.TestProdDatabaseURL, cfg.DatabaseURL)
	assert.Equal(t, testutil.TestEnvProd, cfg.Environment)
}
