package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// SnapshotService manages volume snapshots and storage interactions.
type SnapshotService struct {
	repo       ports.SnapshotRepository
	volumeRepo ports.VolumeRepository
	storage    ports.StorageBackend
	eventSvc   ports.EventService
	auditSvc   ports.AuditService
	logger     *slog.Logger
}

const snapshotNamePrefix = "thecloud-snap-"
const volumeNamePrefix = "thecloud-vol-"

// NewSnapshotService constructs a SnapshotService with its dependencies.
func NewSnapshotService(
	repo ports.SnapshotRepository,
	volumeRepo ports.VolumeRepository,
	storage ports.StorageBackend,
	eventSvc ports.EventService,
	auditSvc ports.AuditService,
	logger *slog.Logger,
) *SnapshotService {
	return &SnapshotService{
		repo:       repo,
		volumeRepo: volumeRepo,
		storage:    storage,
		eventSvc:   eventSvc,
		auditSvc:   auditSvc,
		logger:     logger,
	}
}

func (s *SnapshotService) CreateSnapshot(ctx context.Context, volumeID uuid.UUID, description string) (*domain.Snapshot, error) {
	// 1. Get volume
	vol, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return nil, err
	}

	// 2. Create domain entity
	snapshot := &domain.Snapshot{
		ID:          uuid.New(),
		UserID:      appcontext.UserIDFromContext(ctx),
		VolumeID:    volumeID,
		VolumeName:  vol.Name,
		SizeGB:      vol.SizeGB,
		Status:      domain.SnapshotStatusCreating,
		Description: description,
		CreatedAt:   time.Now(),
	}

	// 3. Persist to DB initially
	if err := s.repo.Create(ctx, snapshot); err != nil {
		return nil, err
	}

	// 4. Perform async snapshot
	// Copy snapshot to avoid data race with returned pointer
	asyncSnap := *snapshot
	go func() {
		bgCtx := context.Background()
		err := s.performSnapshot(bgCtx, vol, &asyncSnap)
		if err != nil {
			s.logger.Error("failed to perform snapshot", "snapshot_id", snapshot.ID, "error", err)
			asyncSnap.Status = domain.SnapshotStatusError
		} else {
			asyncSnap.Status = domain.SnapshotStatusAvailable
		}
		_ = s.repo.Update(bgCtx, &asyncSnap)
	}()

	_ = s.eventSvc.RecordEvent(ctx, "SNAPSHOT_CREATE", snapshot.ID.String(), "SNAPSHOT", map[string]interface{}{
		"volume_id": volumeID.String(),
	})

	_ = s.auditSvc.Log(ctx, snapshot.UserID, "snapshot.create", "snapshot", snapshot.ID.String(), map[string]interface{}{
		"volume_id": volumeID.String(),
	})

	return snapshot, nil
}

func (s *SnapshotService) ListSnapshots(ctx context.Context) ([]*domain.Snapshot, error) {
	userID := appcontext.UserIDFromContext(ctx)
	return s.repo.ListByUserID(ctx, userID)
}

func (s *SnapshotService) GetSnapshot(ctx context.Context, id uuid.UUID) (*domain.Snapshot, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SnapshotService) performSnapshot(ctx context.Context, vol *domain.Volume, snapshot *domain.Snapshot) error {
	backendVolName := volumeNamePrefix + vol.ID.String()[:8]
	backendSnapName := snapshotNamePrefix + snapshot.ID.String()[:8]

	if err := s.storage.CreateSnapshot(ctx, backendVolName, backendSnapName); err != nil {
		return fmt.Errorf("failed to create block volume snapshot: %w", err)
	}

	return nil
}

func (s *SnapshotService) DeleteSnapshot(ctx context.Context, id uuid.UUID) error {
	snapshot, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 1. Delete from Backend
	backendSnapName := snapshotNamePrefix + snapshot.ID.String()[:8]
	if err := s.storage.DeleteSnapshot(ctx, backendSnapName); err != nil {
		s.logger.Warn("failed to delete backend snapshot", "name", backendSnapName, "error", err)
	}

	// 2. Delete from DB
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.eventSvc.RecordEvent(ctx, "SNAPSHOT_DELETE", id.String(), "SNAPSHOT", map[string]interface{}{})
	_ = s.auditSvc.Log(ctx, snapshot.UserID, "snapshot.delete", "snapshot", id.String(), map[string]interface{}{})

	return nil
}

func (s *SnapshotService) RestoreSnapshot(ctx context.Context, snapshotID uuid.UUID, newVolumeName string) (*domain.Volume, error) {
	snapshot, err := s.repo.GetByID(ctx, snapshotID)
	if err != nil {
		return nil, err
	}

	if snapshot.Status != domain.SnapshotStatusAvailable {
		return nil, errors.New(errors.InvalidInput, "cannot restore from snapshot that is not available")
	}

	// 1. Create new volume domain entity
	vol := &domain.Volume{
		ID:        uuid.New(),
		UserID:    snapshot.UserID,
		Name:      newVolumeName,
		SizeGB:    snapshot.SizeGB,
		Status:    domain.VolumeStatusAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 2. Create Block Volume
	backendVolName := volumeNamePrefix + vol.ID.String()[:8]
	path, err := s.storage.CreateVolume(ctx, backendVolName, vol.SizeGB)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create volume for restore", err)
	}
	vol.BackendPath = path

	// 3. Restore snapshot into new volume
	backendSnapName := snapshotNamePrefix + snapshot.ID.String()[:8]
	if err := s.storage.RestoreSnapshot(ctx, backendVolName, backendSnapName); err != nil {
		_ = s.storage.DeleteVolume(ctx, backendVolName)
		return nil, fmt.Errorf("failed to restore volume snapshot: %w", err)
	}

	// 4. Save to DB
	if err := s.volumeRepo.Create(ctx, vol); err != nil {
		_ = s.storage.DeleteVolume(ctx, backendVolName)
		return nil, err
	}

	_ = s.eventSvc.RecordEvent(ctx, "VOLUME_RESTORE", vol.ID.String(), "VOLUME", map[string]interface{}{
		"snapshot_id": snapshotID.String(),
	})

	_ = s.auditSvc.Log(ctx, vol.UserID, "volume.restore", "volume", vol.ID.String(), map[string]interface{}{
		"snapshot_id": snapshotID.String(),
	})

	return vol, nil
}
