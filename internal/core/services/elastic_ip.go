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

// ElasticIPServiceParams holds the dependencies for creating an ElasticIPService.
type ElasticIPServiceParams struct {
	Repo         ports.ElasticIPRepository
	InstanceRepo ports.InstanceRepository
	AuditSvc     ports.AuditService
	Logger       *slog.Logger
}

// NewElasticIPService constructs an ElasticIPService using the provided parameters.
func NewElasticIPService(params ElasticIPServiceParams) ports.ElasticIPService {
	return &elasticIPService{
		repo:         params.Repo,
		instanceRepo: params.InstanceRepo,
		auditSvc:     params.AuditSvc,
		logger:       params.Logger,
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

	if err := s.auditSvc.Log(ctx, userID, "eip.allocate", "eip", id.String(), map[string]interface{}{
		"public_ip": publicIP,
	}); err != nil {
		s.logger.Warn("audit log failed for eip.allocate", "error", err)
	}

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

	if err := s.auditSvc.Log(ctx, eip.UserID, "eip.release", "eip", id.String(), map[string]interface{}{
		"public_ip": eip.PublicIP,
	}); err != nil {
		s.logger.Warn("audit log failed for eip.release", "error", err)
	}

	return nil
}

func (s *elasticIPService) AssociateIP(ctx context.Context, id uuid.UUID, instanceID uuid.UUID) (*domain.ElasticIP, error) {
	eip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
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
	existing, err := s.repo.GetByInstanceID(ctx, instanceID)
	if err != nil && !errors.Is(err, errors.NotFound) {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, errors.New(errors.Conflict, "instance already has an associated elastic ip")
	}

	// 2.1 Verify EIP status
	if eip.Status != domain.EIPStatusAllocated {
		// If re-associating same instance, idempotent success
		if eip.InstanceID != nil && *eip.InstanceID == instanceID {
			return eip, nil
		}
		return nil, errors.New(errors.Conflict, "elastic ip is already associated with another instance")
	}

	// 3. Update EIP mapping
	eip.InstanceID = &instanceID
	eip.VpcID = inst.VpcID
	eip.Status = domain.EIPStatusAssociated
	eip.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, eip); err != nil {
		return nil, err
	}

	if err := s.auditSvc.Log(ctx, eip.UserID, "eip.associate", "eip", id.String(), map[string]interface{}{
		"instance_id": instanceID,
		"public_ip":   eip.PublicIP,
	}); err != nil {
		s.logger.Warn("audit log failed for eip.associate", "error", err)
	}

	return eip, nil
}

func (s *elasticIPService) DisassociateIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error) {
	eip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
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

	if err := s.auditSvc.Log(ctx, eip.UserID, "eip.disassociate", "eip", id.String(), map[string]interface{}{
		"instance_id": oldInstanceID,
		"public_ip":   eip.PublicIP,
	}); err != nil {
		s.logger.Warn("audit log failed for eip.disassociate", "error", err)
	}

	return eip, nil
}

func (s *elasticIPService) ListElasticIPs(ctx context.Context) ([]*domain.ElasticIP, error) {
	return s.repo.List(ctx)
}

func (s *elasticIPService) GetElasticIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error) {
	return s.repo.GetByID(ctx, id)
}

const (
	// CGNAT_FIRST_OCTET is the first octet of the Carrier-Grade NAT range (100.64.0.0/10).
	CGNAT_FIRST_OCTET = 100
	// CGNAT_SECOND_OCTET_BASE is the starting second octet of the CGNAT range.
	CGNAT_SECOND_OCTET_BASE = 64
	// CGNAT_SECOND_OCTET_MASK is the number of addresses in the /10 range (within the second octet).
	CGNAT_SECOND_OCTET_MASK = 64

	// UUID byte indices used for deterministic IP generation.
	UUID_IP_BYTE_1 = 12
	// UUID_IP_BYTE_2 is the second byte index from UUID.
	UUID_IP_BYTE_2 = 13
	// UUID_IP_BYTE_3 is the third byte index from UUID.
	UUID_IP_BYTE_3 = 14
)

// generateDeterministicIP creates a consistent "public" IP for a given UUID within the 100.64.0.0/10 range.
func (s *elasticIPService) generateDeterministicIP(u uuid.UUID) string {
	// 100.64.0.0 + bytes 12-15 of UUID
	ip := net.IPv4(CGNAT_FIRST_OCTET, CGNAT_SECOND_OCTET_BASE+u[UUID_IP_BYTE_1]%CGNAT_SECOND_OCTET_MASK, u[UUID_IP_BYTE_2], u[UUID_IP_BYTE_3])
	return ip.String()
}
