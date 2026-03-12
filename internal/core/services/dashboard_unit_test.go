package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestDashboardService_GetStats(t *testing.T) {
	instRepo := new(MockInstanceRepo)
	volRepo := new(MockVolumeRepo)
	vpcRepo := new(MockVpcRepo)
	eventRepo := new(MockEventRepo)
	rbacSvc := new(MockRBACService)

	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	svc := services.NewDashboardService(rbacSvc, instRepo, volRepo, vpcRepo, eventRepo, slog.Default())
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		instID := uuid.New()
		instances := []*domain.Instance{
			{ID: uuid.New(), Status: domain.StatusRunning},
			{ID: uuid.New(), Status: domain.StatusStopped},
		}
		volumes := []*domain.Volume{
			{ID: uuid.New(), SizeGB: 10, InstanceID: &instID},
			{ID: uuid.New(), SizeGB: 20},
		}
		vpcs := []*domain.VPC{{ID: uuid.New()}, {ID: uuid.New()}, {ID: uuid.New()}}
		events := []*domain.Event{{ID: uuid.New(), Action: "test.action"}}

		rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		instRepo.On("List", ctx).Return(instances, nil).Once()
		volRepo.On("List", ctx).Return(volumes, nil).Once()
		vpcRepo.On("List", ctx).Return(vpcs, nil).Once()
		eventRepo.On("List", ctx, 10).Return(events, nil).Once()

		stats, err := svc.GetStats(ctx)
		require.NoError(t, err)
		assert.NotNil(t, stats)

		assert.Equal(t, 2, stats.Summary.TotalInstances)
		assert.Equal(t, 1, stats.Summary.RunningInstances)
		assert.Equal(t, 1, stats.Summary.StoppedInstances)
		assert.Equal(t, 2, stats.Summary.TotalVolumes)
		assert.Equal(t, 1, stats.Summary.AttachedVolumes)
		assert.Equal(t, 30*1024, stats.Summary.TotalStorageMB)
		assert.Equal(t, 3, stats.Summary.TotalVPCs)
		assert.Len(t, stats.RecentEvents, 1)
	})
}
