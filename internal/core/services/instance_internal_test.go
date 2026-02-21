package services

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testVolumeName = "test-vol"

// mockVolumeRepo is already defined in dashboard_test.go (package services)

func TestInstanceServiceInternalGetVolumeByIDOrName(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	volID := uuid.New()

	t.Run("ByID", func(t *testing.T) {
		repo.On("GetByID", ctx, volID).Return(&domain.Volume{ID: volID}, nil).Once()
		res, err := svc.getVolumeByIDOrName(ctx, volID.String())
		assert.NoError(t, err)
		assert.Equal(t, volID, res.ID)
	})

	t.Run("ByName", func(t *testing.T) {
		repo.On("GetByName", ctx, testVolumeName).Return(&domain.Volume{Name: testVolumeName}, nil).Once()
		res, err := svc.getVolumeByIDOrName(ctx, testVolumeName)
		assert.NoError(t, err)
		assert.Equal(t, testVolumeName, res.Name)
	})
}

func TestInstanceServiceInternalResolveVolumes(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	volID := uuid.New()

	repo.On("GetByID", ctx, volID).Return(&domain.Volume{ID: volID, Name: "vol1", Status: domain.VolumeStatusAvailable}, nil).Once()

	binds, vols, err := svc.resolveVolumes(ctx, []domain.VolumeAttachment{{VolumeIDOrName: volID.String(), MountPath: "/data"}})
	assert.NoError(t, err)
	assert.Len(t, binds, 1)
	assert.Len(t, vols, 1)
}

func TestInstanceServiceInternalResolveVolumesUnavailable(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	volID := uuid.New()

	repo.On("GetByID", ctx, volID).Return(&domain.Volume{ID: volID, Name: "vol1", Status: domain.VolumeStatusInUse}, nil).Once()

	_, _, err := svc.resolveVolumes(ctx, []domain.VolumeAttachment{{VolumeIDOrName: volID.String(), MountPath: "/data"}})
	assert.Error(t, err)
}

func TestInstanceServiceInternalUpdateVolumesAfterLaunch(t *testing.T) {
	t.Parallel()
	repo := new(mockVolumeRepo)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &InstanceService{volumeRepo: repo, logger: logger}
	ctx := context.Background()
	instID := uuid.New()
	vol := &domain.Volume{ID: uuid.New(), Status: domain.VolumeStatusAvailable}

	repo.On("Update", ctx, mock.MatchedBy(func(v *domain.Volume) bool {
		return v.Status == domain.VolumeStatusInUse && v.InstanceID != nil && *v.InstanceID == instID
	})).Return(nil).Once()

	svc.updateVolumesAfterLaunch(ctx, []*domain.Volume{vol}, instID)
	repo.AssertExpectations(t)
}

func TestInstanceService_CalculateInstanceStats(t *testing.T) {
	svc := &InstanceService{}
	stats := &domain.RawDockerStats{}
	
	stats.CPUStats.CPUUsage.TotalUsage = 1000
	stats.CPUStats.SystemCPUUsage = 10000
	
	stats.PreCPUStats.CPUUsage.TotalUsage = 500
	stats.PreCPUStats.SystemCPUUsage = 5000
	
	stats.MemoryStats.Usage = 1024
	stats.MemoryStats.Limit = 2048

	res := svc.calculateInstanceStats(stats)
	assert.InDelta(t, 10.0, res.CPUPercentage, 0.01) // (1000-500)/(10000-5000) * 100 = 10%
	assert.InDelta(t, 50.0, res.MemoryPercentage, 0.01)
}

func TestInstanceService_FormatContainerName(t *testing.T) {
	svc := &InstanceService{}
	id := uuid.New()
	name := svc.formatContainerName(id)
	assert.Equal(t, "thecloud-"+id.String()[:8], name)
}

func TestInstanceService_IsValidHostIP(t *testing.T) {
	svc := &InstanceService{}
	_, ipNet, _ := net.ParseCIDR("10.0.0.0/24")

	t.Run("Valid", func(t *testing.T) {
		assert.True(t, svc.isValidHostIP(net.ParseIP("10.0.0.5"), ipNet))
	})

	t.Run("NetworkAddress", func(t *testing.T) {
		assert.False(t, svc.isValidHostIP(net.ParseIP("10.0.0.0"), ipNet))
	})

	t.Run("BroadcastAddress", func(t *testing.T) {
		assert.False(t, svc.isValidHostIP(net.ParseIP("10.0.0.255"), ipNet))
	})

	t.Run("OutsideSubnet", func(t *testing.T) {
		assert.False(t, svc.isValidHostIP(net.ParseIP("10.0.1.5"), ipNet))
	})
}

func TestParsePort(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := parsePort("80")
		assert.NoError(t, err)
		assert.Equal(t, 80, p)
	})

	t.Run("Empty", func(t *testing.T) {
		_, err := parsePort("")
		assert.Error(t, err)
	})

	t.Run("Invalid", func(t *testing.T) {
		_, err := parsePort("abc")
		assert.Error(t, err)
	})
}

func TestInstanceService_UpdateInstanceMetadata(t *testing.T) {
	repo := new(mockInstanceRepo)
	svc := &InstanceService{repo: repo}
	ctx := context.Background()
	id := uuid.New()
	inst := &domain.Instance{
		ID:       id,
		Metadata: map[string]string{"old": "val"},
		Labels:   map[string]string{"l1": "v1"},
	}

	repo.On("GetByID", ctx, id).Return(inst, nil).Once()
	repo.On("Update", ctx, inst).Return(nil).Once()

	metadata := map[string]string{"new": "val", "old": ""} // "old" should be deleted
	labels := map[string]string{"l2": "v2"}

	err := svc.UpdateInstanceMetadata(ctx, id, metadata, labels)
	assert.NoError(t, err)
	assert.Equal(t, "val", inst.Metadata["new"])
	assert.Equal(t, "v2", inst.Labels["l2"])
	assert.Equal(t, "v1", inst.Labels["l1"])
	_, ok := inst.Metadata["old"]
	assert.False(t, ok)
}
