package services

import (
	"context"
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
	// Note: In a real "Async by Default" architecture, this would be a background job.
	// For now, we do it synchronously to see results.
	containerID, err := s.docker.CreateContainer(ctx, name, image)
	_ = containerID // ID is not yet stored, but container is created
	if err != nil {
		inst.Status = domain.StatusError
		_ = s.repo.Update(ctx, inst) // Try to mark error in DB
		return nil, errors.Wrap(errors.Internal, "failed to launch container", err)
	}

	// 4. Update status to RUNNING
	inst.Status = domain.StatusRunning
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
	// In the real world, instance ID would map to container ID.
	// We'll use the instance name or a stored ID for this.
	if err := s.docker.StopContainer(ctx, inst.Name); err != nil {
		return errors.Wrap(errors.Internal, "failed to stop container", err)
	}

	// 3. Update DB
	inst.Status = domain.StatusStopped
	return s.repo.Update(ctx, inst)
}

func (s *InstanceService) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	return s.repo.List(ctx)
}
