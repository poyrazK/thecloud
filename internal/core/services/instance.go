package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/poyraz/cloud/internal/core/domain"
	"github.com/poyraz/cloud/internal/core/ports"
	"github.com/poyraz/cloud/internal/errors"
)

type InstanceService struct {
	repo   ports.InstanceRepository
	docker ports.DockerClient
}

func NewInstanceService(repo ports.InstanceRepository, docker ports.DockerClient) *InstanceService {
	return &InstanceService{
		repo:   repo,
		docker: docker,
	}
}

func (s *InstanceService) LaunchInstance(ctx context.Context, name, image string) (*domain.Instance, error) {
	// 1. Create domain entity
	inst := &domain.Instance{
		ID:        uuid.New(),
		Name:      name,
		Image:     image,
		Status:    domain.StatusStarting,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 2. Persist to DB first (Pending state)
	if err := s.repo.Create(ctx, inst); err != nil {
		return nil, err
	}

	// 3. Call Docker to create actual container
	// We generate a unique name for Docker to avoid conflicts if user reuses "Name"
	// Docker name format: miniaws-<short_uuid>
	dockerName := fmt.Sprintf("miniaws-%s", inst.ID.String()[:8])

	containerID, err := s.docker.CreateContainer(ctx, dockerName, image)
	if err != nil {
		inst.Status = domain.StatusError
		_ = s.repo.Update(ctx, inst) // Try to mark error in DB
		return nil, errors.Wrap(errors.Internal, "failed to launch container", err)
	}

	// 4. Update status and save ContainerID
	inst.Status = domain.StatusRunning
	inst.ContainerID = containerID
	// Note: We might want to store the docker container ID in our DB too.
	// For simplicity, we'll just update status for now.
	if err := s.repo.Update(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

func (s *InstanceService) StopInstance(ctx context.Context, id uuid.UUID) error {
	// 1. Get from DB
	inst, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if inst.Status == domain.StatusStopped {
		return nil // Already stopped
	}

	// 2. Call Docker stop
	// We use the stored ContainerID if available, otherwise fallback to Name (legacy support)
	target := inst.ContainerID
	if target == "" {
		// Try to reconstruct older naming scheme or just fail?
		// For now let's assumes legacy containers used Name.
		// But in our new scheme they use miniaws-ID.
		// Actually, let's look for the container by name if ID is missing.
		target = inst.Name
	} else {
		// StopContainer in adapter currently takes 'name', but Docker API supports ID too.
	}

	if err := s.docker.StopContainer(ctx, target); err != nil {
		// If we can't stop it, maybe it is already gone or name changed.
		// Log warning?
		return errors.Wrap(errors.Internal, "failed to stop container", err)
	}

	// 3. Update DB
	inst.Status = domain.StatusStopped
	return s.repo.Update(ctx, inst)
}

func (s *InstanceService) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	return s.repo.List(ctx)
}
