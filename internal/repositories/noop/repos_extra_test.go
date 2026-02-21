package noop

import (
	"context"
	"testing"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestNoopRepositoriesExtra(t *testing.T) {
	ctx := context.Background()
	uID := uuid.New()
	id := uuid.New()

	t.Run("InstanceRepository", func(t *testing.T) {
		repo := &NoopInstanceRepository{}
		list, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)

		list, err = repo.ListAll(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)

		assert.NoError(t, repo.Update(ctx, &domain.Instance{}))

		list, err = repo.ListBySubnet(ctx, id)
		assert.NoError(t, err)
		assert.Empty(t, list)

		list, err = repo.ListByUserID(ctx, uID)
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("VpcRepository", func(t *testing.T) {
		repo := &NoopVpcRepository{}
		assert.NoError(t, repo.Create(ctx, &domain.VPC{}))
		_, err := repo.GetByName(ctx, "name")
		assert.NoError(t, err)
		list, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)
		list, err = repo.ListByUserID(ctx, uID)
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("SubnetRepository", func(t *testing.T) {
		repo := &NoopSubnetRepository{}
		assert.NoError(t, repo.Create(ctx, &domain.Subnet{}))
		list, err := repo.ListByVPC(ctx, id)
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("VolumeRepository", func(t *testing.T) {
		repo := &NoopVolumeRepository{}
		assert.NoError(t, repo.Create(ctx, &domain.Volume{}))
		_, err := repo.GetByName(ctx, "name")
		assert.NoError(t, err)
		list, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)
		list, err = repo.ListByUserID(ctx, uID)
		assert.NoError(t, err)
		assert.Empty(t, list)
		list, err = repo.ListByInstanceID(ctx, id)
		assert.NoError(t, err)
		assert.Empty(t, list)
		assert.NoError(t, repo.Update(ctx, &domain.Volume{}))
	})

	t.Run("DatabaseRepository", func(t *testing.T) {
		repo := &NoopDatabaseRepository{}
		assert.NoError(t, repo.Create(ctx, &domain.Database{}))
		_, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		list, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)
		assert.NoError(t, repo.Update(ctx, &domain.Database{}))
		assert.NoError(t, repo.Delete(ctx, id))
	})

	t.Run("CacheRepository", func(t *testing.T) {
		repo := &NoopCacheRepository{}
		assert.NoError(t, repo.Create(ctx, &domain.Cache{}))
		_, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		_, err = repo.GetByName(ctx, uID, "name")
		assert.NoError(t, err)
		list, err := repo.List(ctx, uID)
		assert.NoError(t, err)
		assert.Empty(t, list)
		assert.NoError(t, repo.Update(ctx, &domain.Cache{}))
		assert.NoError(t, repo.Delete(ctx, id))
	})

	t.Run("FunctionRepository", func(t *testing.T) {
		repo := &NoopFunctionRepository{}
		assert.NoError(t, repo.Create(ctx, &domain.Function{}))
		_, err := repo.GetByName(ctx, uID, "name")
		assert.NoError(t, err)
		list, err := repo.List(ctx, uID)
		assert.NoError(t, err)
		assert.Empty(t, list)
		assert.NoError(t, repo.Delete(ctx, id))
	})

	t.Run("UserRepository", func(t *testing.T) {
		repo := &NoopUserRepository{}
		require.NoError(t, repo.Create(ctx, &domain.User{}))
		user, err := repo.GetByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.Nil(t, user)
		_, err = repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, repo.Update(ctx, &domain.User{}))
		list, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Empty(t, list)
		assert.NoError(t, repo.Delete(ctx, id))
	})
}
