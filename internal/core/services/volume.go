// Package services implements core business workflows.
package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const tracerName = "volume-service"

// VolumeService manages block volume lifecycle and attachments.
type VolumeService struct {
	repo     ports.VolumeRepository
	storage  ports.StorageBackend
	eventSvc ports.EventService
	auditSvc ports.AuditService
	logger   *slog.Logger
}

// NewVolumeService constructs a VolumeService with its dependencies.
func NewVolumeService(repo ports.VolumeRepository, storage ports.StorageBackend, eventSvc ports.EventService, auditSvc ports.AuditService, logger *slog.Logger) *VolumeService {
	return &VolumeService{
		repo:     repo,
		storage:  storage,
		eventSvc: eventSvc,
		auditSvc: auditSvc,
		logger:   logger,
	}
}

func (s *VolumeService) CreateVolume(ctx context.Context, name string, sizeGB int) (*domain.Volume, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "CreateVolume")
	defer span.End()

	span.SetAttributes(
		attribute.String("volume.name", name),
		attribute.Int("volume.size_gb", sizeGB),
	)
	// 1. Create domain entity
	vol := &domain.Volume{
		ID:        uuid.New(),
		UserID:    appcontext.UserIDFromContext(ctx),
		Name:      name,
		SizeGB:    sizeGB,
		Status:    domain.VolumeStatusAvailable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 2. Create Block Volume
	backendName := "thecloud-vol-" + vol.ID.String()[:8]
	path, err := s.storage.CreateVolume(ctx, backendName, sizeGB)
	if err != nil {
		s.logger.Error("failed to create storage volume", "name", backendName, "error", err)
		return nil, errors.Wrap(errors.Internal, "failed to create volume", err)
	}
	vol.BackendPath = path // We need to add this field to domain.Volume

	// 3. Persist to DB
	if err := s.repo.Create(ctx, vol); err != nil {
		// Rollback Storage Volume
		_ = s.storage.DeleteVolume(ctx, backendName)
		return nil, err
	}

	_ = s.eventSvc.RecordEvent(ctx, "VOLUME_CREATE", vol.ID.String(), "VOLUME", map[string]interface{}{
		"name":    vol.Name,
		"size_gb": vol.SizeGB,
		"path":    path,
	})

	_ = s.auditSvc.Log(ctx, vol.UserID, "volume.create", "volume", vol.ID.String(), map[string]interface{}{
		"name":    vol.Name,
		"size_gb": vol.SizeGB,
	})

	s.logger.Info("volume created", "volume_id", vol.ID, "name", vol.Name, "path", path)
	// platform metrics ...
	return vol, nil
}

func (s *VolumeService) ListVolumes(ctx context.Context) ([]*domain.Volume, error) {
	return s.repo.List(ctx)
}

func (s *VolumeService) GetVolume(ctx context.Context, idOrName string) (*domain.Volume, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "GetVolume")
	defer span.End()
	span.SetAttributes(attribute.String("volume.id_or_name", idOrName))
	id, err := uuid.Parse(idOrName)
	if err == nil {
		return s.repo.GetByID(ctx, id)
	}
	return s.repo.GetByName(ctx, idOrName)
}

func (s *VolumeService) DeleteVolume(ctx context.Context, idOrName string) error {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "DeleteVolume")
	defer span.End()
	span.SetAttributes(attribute.String("volume.id_or_name", idOrName))
	vol, err := s.GetVolume(ctx, idOrName)
	if err != nil {
		return err
	}

	if vol.Status == domain.VolumeStatusInUse {
		return errors.New(errors.InvalidInput, "cannot delete volume that is in use")
	}

	// 1. Delete Storage Volume
	backendName := "thecloud-vol-" + vol.ID.String()[:8]
	if err := s.storage.DeleteVolume(ctx, backendName); err != nil {
		s.logger.Warn("failed to delete storage volume", "name", backendName, "error", err)
	}

	// 2. Delete from DB
	if err := s.repo.Delete(ctx, vol.ID); err != nil {
		return err
	}

	platform.VolumesTotal.WithLabelValues(string(vol.Status)).Dec()
	platform.VolumeSizeBytes.Sub(float64(vol.SizeGB * 1024 * 1024 * 1024))
	platform.StorageOperationsTotal.WithLabelValues("volume_delete").Inc()

	_ = s.eventSvc.RecordEvent(ctx, "VOLUME_DELETE", vol.ID.String(), "VOLUME", map[string]interface{}{})

	_ = s.auditSvc.Log(ctx, vol.UserID, "volume.delete", "volume", vol.ID.String(), map[string]interface{}{
		"name": vol.Name,
	})

	s.logger.Info("volume deleted", "volume_id", vol.ID)
	return nil
}

// ReleaseVolumesForInstance detaches all volumes attached to an instance and marks them as available.
// This should be called when an instance is terminated to free up its volumes.
func (s *VolumeService) ReleaseVolumesForInstance(ctx context.Context, instanceID uuid.UUID) error {
	volumes, err := s.repo.ListByInstanceID(ctx, instanceID)
	if err != nil {
		s.logger.Error("failed to list volumes for instance", "instance_id", instanceID, "error", err)
		return err
	}

	for _, vol := range volumes {
		vol.Status = domain.VolumeStatusAvailable
		vol.InstanceID = nil
		vol.MountPath = ""
		vol.UpdatedAt = time.Now()

		if err := s.repo.Update(ctx, vol); err != nil {
			s.logger.Warn("failed to release volume", "volume_id", vol.ID, "error", err)
			continue // Continue releasing other volumes even if one fails
		}

		s.logger.Info("volume released", "volume_id", vol.ID, "instance_id", instanceID)
	}

	return nil
}
