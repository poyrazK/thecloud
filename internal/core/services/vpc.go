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
)

type VpcService struct {
	repo     ports.VpcRepository
	docker   ports.DockerClient
	auditSvc ports.AuditService
	logger   *slog.Logger
}

func NewVpcService(repo ports.VpcRepository, docker ports.DockerClient, auditSvc ports.AuditService, logger *slog.Logger) *VpcService {
	return &VpcService{
		repo:     repo,
		docker:   docker,
		auditSvc: auditSvc,
		logger:   logger,
	}
}

func (s *VpcService) CreateVPC(ctx context.Context, name string) (*domain.VPC, error) {
	// 1. Create Docker network first
	networkName := fmt.Sprintf("thecloud-vpc-%s", uuid.New().String()[:8])
	dockerNetworkID, err := s.docker.CreateNetwork(ctx, networkName)
	if err != nil {
		return nil, err
	}

	// 2. Persist to DB
	vpc := &domain.VPC{
		ID:        uuid.New(),
		UserID:    appcontext.UserIDFromContext(ctx),
		Name:      name,
		NetworkID: dockerNetworkID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, vpc); err != nil {
		// Cleanup Docker network if DB fails
		s.logger.Error("failed to create VPC in DB, rolling back network", "name", name, "error", err)
		if rbErr := s.docker.RemoveNetwork(ctx, dockerNetworkID); rbErr != nil {
			s.logger.Error("failed to rollback network", "network_id", dockerNetworkID, "error", rbErr)
		}
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, vpc.UserID, "vpc.create", "vpc", vpc.ID.String(), map[string]interface{}{
		"name": vpc.Name,
	})

	return vpc, nil
}

func (s *VpcService) GetVPC(ctx context.Context, idOrName string) (*domain.VPC, error) {
	id, err := uuid.Parse(idOrName)
	if err == nil {
		return s.repo.GetByID(ctx, id)
	}
	return s.repo.GetByName(ctx, idOrName)
}

func (s *VpcService) ListVPCs(ctx context.Context) ([]*domain.VPC, error) {
	return s.repo.List(ctx)
}

func (s *VpcService) DeleteVPC(ctx context.Context, idOrName string) error {
	vpc, err := s.GetVPC(ctx, idOrName)
	if err != nil {
		return err
	}

	// 1. Remove Docker network
	if err := s.docker.RemoveNetwork(ctx, vpc.NetworkID); err != nil {
		s.logger.Error("failed to remove docker network", "network_id", vpc.NetworkID, "error", err)
		return err
	}
	s.logger.Info("vpc network removed", "network_id", vpc.NetworkID)

	// 2. Delete from DB
	if err := s.repo.Delete(ctx, vpc.ID); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, vpc.UserID, "vpc.delete", "vpc", vpc.ID.String(), map[string]interface{}{
		"name": vpc.Name,
	})

	return nil
}
