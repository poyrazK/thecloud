package services

import (
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseService_Internal(t *testing.T) {
	s := &DatabaseService{}

	t.Run("parseAllocatedPort", func(t *testing.T) {
		port, err := s.parseAllocatedPort([]string{"30001:5432"}, "5432")
		assert.NoError(t, err)
		assert.Equal(t, 30001, port)

		port, _ = s.parseAllocatedPort([]string{"80:80"}, "5432")
		assert.Equal(t, 0, port)
	})

	t.Run("getEngineConfig", func(t *testing.T) {
		image, env, port := s.getEngineConfig(domain.EnginePostgres, "16", "user", "pass", "mydb", domain.RolePrimary, "")
		assert.Equal(t, "postgres:16-alpine", image)
		assert.Contains(t, env, "POSTGRES_USER=user")
		assert.Equal(t, "5432", port)

		image, _, port = s.getEngineConfig(domain.EngineMySQL, "8.0", "user", "pass", "mydb", domain.RolePrimary, "")
		assert.Equal(t, "mysql:8.0", image)
		assert.Equal(t, "3306", port)
	})
}
