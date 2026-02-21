package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestAutoScalingRepo_Extra(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("GroupInstances", func(t *testing.T) {
		t.Parallel()
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewAutoScalingRepo(mock)
		groupID := uuid.New()
		instanceID := uuid.New()

		// AddInstanceToGroup
		mock.ExpectExec("INSERT INTO scaling_group_instances").
			WithArgs(groupID, instanceID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		err := repo.AddInstanceToGroup(ctx, groupID, instanceID)
		assert.NoError(t, err)

		// GetInstancesInGroup
		mock.ExpectQuery("SELECT instance_id FROM scaling_group_instances").
			WithArgs(groupID).
			WillReturnRows(pgxmock.NewRows([]string{"instance_id"}).AddRow(instanceID))
		ids, err := repo.GetInstancesInGroup(ctx, groupID)
		assert.NoError(t, err)
		assert.Len(t, ids, 1)
		assert.Equal(t, instanceID, ids[0])

		// RemoveInstanceFromGroup
		mock.ExpectExec("DELETE FROM scaling_group_instances").
			WithArgs(groupID, instanceID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		err = repo.RemoveInstanceFromGroup(ctx, groupID, instanceID)
		assert.NoError(t, err)
	})

	t.Run("GetAllScalingGroupInstances", func(t *testing.T) {
		t.Parallel()
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewAutoScalingRepo(mock)
		groupID := uuid.New()
		instanceID := uuid.New()

		mock.ExpectQuery("SELECT scaling_group_id, instance_id FROM scaling_group_instances").
			WithArgs([]uuid.UUID{groupID}).
			WillReturnRows(pgxmock.NewRows([]string{"scaling_group_id", "instance_id"}).AddRow(groupID, instanceID))
		
		result, err := repo.GetAllScalingGroupInstances(ctx, []uuid.UUID{groupID})
		assert.NoError(t, err)
		assert.Len(t, result[groupID], 1)
		assert.Equal(t, instanceID, result[groupID][0])
	})

	t.Run("GetAverageCPU", func(t *testing.T) {
		t.Parallel()
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewAutoScalingRepo(mock)
		instanceID := uuid.New()
		since := time.Now().Add(-time.Hour)

		mock.ExpectQuery("SELECT COALESCE").
			WithArgs([]uuid.UUID{instanceID}, since).
			WillReturnRows(pgxmock.NewRows([]string{"avg"}).AddRow(45.5))
		
		avg, err := repo.GetAverageCPU(ctx, []uuid.UUID{instanceID}, since)
		assert.NoError(t, err)
		assert.Equal(t, 45.5, avg)
	})
	
	t.Run("GetAllPolicies", func(t *testing.T) {
		t.Parallel()
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewAutoScalingRepo(mock)
		groupID := uuid.New()
		policyID := uuid.New()

		mock.ExpectQuery("SELECT id, scaling_group_id, name, metric_type, target_value, scale_out_step, scale_in_step, cooldown_sec, last_scaled_at FROM scaling_policies").
			WithArgs([]uuid.UUID{groupID}).
			WillReturnRows(pgxmock.NewRows([]string{"id", "scaling_group_id", "name", "metric_type", "target_value", "scale_out_step", "scale_in_step", "cooldown_sec", "last_scaled_at"}).
				AddRow(policyID, groupID, "p1", "cpu", 70.0, 1, 1, 300, nil))
		
		result, err := repo.GetAllPolicies(ctx, []uuid.UUID{groupID})
		assert.NoError(t, err)
		assert.Len(t, result[groupID], 1)
		assert.Equal(t, policyID, result[groupID][0].ID)
	})
}
