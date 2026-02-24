package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockEventRepo struct {
	mock.Mock
}

func (m *mockEventRepo) Create(ctx context.Context, e *domain.Event) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockEventRepo) List(ctx context.Context, limit int) ([]*domain.Event, error) {
	args := m.Called(ctx, limit)
	r0, _ := args.Get(0).([]*domain.Event)
	return r0, args.Error(1)
}
func (m *mockEventRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Event, error) {
	args := m.Called(ctx, userID, limit)
	r0, _ := args.Get(0).([]*domain.Event)
	return r0, args.Error(1)
}

func TestDashboardService_GetStats(t *testing.T) {
	instRepo := new(MockInstanceRepo)
	volRepo := new(MockVolumeRepo)
	vpcRepo := new(MockVpcRepo)
	eventRepo := new(mockEventRepo)

	svc := services.NewDashboardService(instRepo, volRepo, vpcRepo, eventRepo, slog.Default())
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
