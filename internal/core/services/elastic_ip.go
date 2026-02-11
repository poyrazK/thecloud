// Package services implements core business logic.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

type elasticIPService struct {
	repo         ports.ElasticIPRepository
	instanceRepo ports.InstanceRepository
	auditSvc     ports.AuditService
	logger       *slog.Logger
}

// NewElasticIPService constructs an ElasticIPService.
func NewElasticIPService(repo ports.ElasticIPRepository, instanceRepo ports.InstanceRepository, auditSvc ports.AuditService, logger *slog.Logger) ports.ElasticIPService {
	return &elasticIPService{
		repo:         repo,
		instanceRepo: instanceRepo,
		auditSvc:     auditSvc,
		logger:       logger,
	}
}

func (s *elasticIPService) AllocateIP(ctx context.Context) (*domain.ElasticIP, error) {
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)
	id := uuid.New()

	// Simulate public IP allocation from CGNAT range 100.64.0.0/10 for demo/simulation
	// In a real system, this would come from an IP pool manager or provider SDK
	publicIP := s.generateDeterministicIP(id)

	eip := &domain.ElasticIP{
		ID:        id,
		UserID:    userID,
		TenantID:  tenantID,
		PublicIP:  publicIP,
		Status:    domain.EIPStatusAllocated,
		ARN:       fmt.Sprintf("arn:thecloud:vpc:local:%s:eip/%s", userID, id),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, eip); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, userID, "eip.allocate", "eip", id.String(), map[string]interface{}{
		"public_ip": publicIP,
	})

	return eip, nil
}

func (s *elasticIPService) ReleaseIP(ctx context.Context, id uuid.UUID) error {
	eip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if eip.Status == domain.EIPStatusAssociated {
		return errors.New(errors.Conflict, "cannot release an associated elastic ip; disassociate it first")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, eip.UserID, "eip.release", "eip", id.String(), map[string]interface{}{
		"public_ip": eip.PublicIP,
	})

	return nil
}

func (s *elasticIPService) AssociateIP(ctx context.Context, id uuid.UUID, instanceID uuid.UUID) (*domain.ElasticIP, error) {
	eip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 1. Check if instance exists and is not terminated
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.Status == domain.StatusDeleted {
		return nil, errors.New(errors.InvalidInput, "cannot associate ip to a deleted instance")
	}

	// 2. Check if instance already has an EIP
	existing, _ := s.repo.GetByInstanceID(ctx, instanceID)
	if existing != nil && existing.ID != id {
		return nil, errors.New(errors.Conflict, "instance already has an associated elastic ip")
	}

	// 3. Update EIP mapping
	eip.InstanceID = &instanceID
	eip.VpcID = inst.VpcID
	eip.Status = domain.EIPStatusAssociated
	eip.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, eip); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, eip.UserID, "eip.associate", "eip", id.String(), map[string]interface{}{
		"instance_id": instanceID,
		"public_ip":   eip.PublicIP,
	})

	return eip, nil
}

func (s *elasticIPService) DisassociateIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error) {
	eip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if eip.Status != domain.EIPStatusAssociated {
		return nil, errors.New(errors.InvalidInput, "elastic ip is not associated")
	}

	oldInstanceID := eip.InstanceID

	eip.InstanceID = nil
	eip.VpcID = nil
	eip.Status = domain.EIPStatusAllocated
	eip.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, eip); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, eip.UserID, "eip.disassociate", "eip", id.String(), map[string]interface{}{
		"instance_id": oldInstanceID,
		"public_ip":   eip.PublicIP,
	})

	return eip, nil
}

func (s *elasticIPService) ListElasticIPs(ctx context.Context) ([]*domain.ElasticIP, error) {
	return s.repo.List(ctx)
}

func (s *elasticIPService) GetElasticIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error) {
	return s.repo.GetByID(ctx, id)
}

// generateDeterministicIP creates a consistent "public" IP for a given UUID within the 100.64.0.0/10 range.
func (s *elasticIPService) generateDeterministicIP(u uuid.UUID) string {
	// 100.64.0.0 + bytes 12-15 of UUID
	ip := net.IPv4(100, 64+u[12]%64, u[13], u[14])
	return ip.String()
}
