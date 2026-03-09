package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElasticIPRepositoryUnit(t *testing.T) {
	t.Parallel()

	t.Run("Create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewElasticIPRepository(mock)
		ctx := context.Background()
		eip := &domain.ElasticIP{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			TenantID:  uuid.New(),
			PublicIP:  "1.2.3.4",
			Status:    "allocated",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectExec("INSERT INTO elastic_ips").
			WithArgs(eip.ID, eip.UserID, eip.TenantID, eip.PublicIP, eip.InstanceID, eip.VpcID, eip.Status, eip.ARN, eip.CreatedAt, eip.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(ctx, eip)
		require.NoError(t, err)
	})

	t.Run("GetByID", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		repo := NewElasticIPRepository(mock)
		id := uuid.New()
		tenantID := uuid.New()
		ctx := appcontext.WithTenantID(context.Background(), tenantID)
		now := time.Now()

		mock.ExpectQuery("SELECT .* FROM elastic_ips WHERE id = \\$1 AND tenant_id = \\$2").
			WithArgs(id, tenantID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "tenant_id", "public_ip", "instance_id", "vpc_id", "status", "arn", "created_at", "updated_at"}).
				AddRow(id, uuid.New(), tenantID, "1.2.3.4", nil, nil, "allocated", "arn", now, now))

		res, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, id, res.ID)
	})
}
