// Package services implements core business workflows.
package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// LBService manages load balancers and target registration.
type LBService struct {
	lbRepo       ports.LBRepository
	vpcRepo      ports.VpcRepository
	instanceRepo ports.InstanceRepository
	auditSvc     ports.AuditService
}

// NewLBService constructs an LBService with its dependencies.
func NewLBService(lbRepo ports.LBRepository, vpcRepo ports.VpcRepository, instanceRepo ports.InstanceRepository, auditSvc ports.AuditService) *LBService {
	return &LBService{
		lbRepo:       lbRepo,
		vpcRepo:      vpcRepo,
		instanceRepo: instanceRepo,
		auditSvc:     auditSvc,
	}
}

func (s *LBService) Create(ctx context.Context, name string, vpcID uuid.UUID, port int, algo string, idempotencyKey string) (*domain.LoadBalancer, error) {
	// Check if already created via idempotency key
	if idempotencyKey != "" {
		existing, err := s.lbRepo.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil {
			return existing, nil
		}
	}

	// Validate VPC exists
	_, err := s.vpcRepo.GetByID(ctx, vpcID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "VPC not found", err)
	}

	// Set default algorithm
	if algo == "" {
		algo = "round-robin"
	}

	lb := &domain.LoadBalancer{
		ID:             uuid.New(),
		UserID:         appcontext.UserIDFromContext(ctx),
		IdempotencyKey: idempotencyKey,
		Name:           name,
		VpcID:          vpcID,
		Port:           port,
		Algorithm:      algo,
		Status:         domain.LBStatusCreating,
		Version:        1,
		CreatedAt:      time.Now(),
	}

	if err := s.lbRepo.Create(ctx, lb); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, lb.UserID, "lb.create", "loadbalancer", lb.ID.String(), map[string]interface{}{
		"name": lb.Name,
	})

	return lb, nil
}

func (s *LBService) Get(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	return s.lbRepo.GetByID(ctx, id)
}

func (s *LBService) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	return s.lbRepo.List(ctx)
}

func (s *LBService) Delete(ctx context.Context, id uuid.UUID) error {
	// Mark as deleted first or just delete?
	// Plan says "DRAINING | DELETED". For now let's just delete or mark DELETED.
	// Since we are async, maybe mark as DELETED and let worker cleanup?
	// For simplicity, let's mark as DELETED.
	lb, err := s.lbRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	lb.Status = domain.LBStatusDeleted
	if err := s.lbRepo.Update(ctx, lb); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, lb.UserID, "lb.delete", "loadbalancer", lb.ID.String(), map[string]interface{}{
		"name": lb.Name,
	})

	return nil
}

func (s *LBService) AddTarget(ctx context.Context, lbID, instanceID uuid.UUID, port int, weight int) error {
	// Get LB
	lb, err := s.lbRepo.GetByID(ctx, lbID)
	if err != nil {
		return err
	}

	// Get Instance
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return errors.Wrap(errors.NotFound, "target instance not found", err)
	}

	// Validate cross-VPC
	if inst.VpcID == nil || *inst.VpcID != lb.VpcID {
		return errors.ErrLBCrossVPC
	}

	if weight == 0 {
		weight = 1
	}

	target := &domain.LBTarget{
		ID:         uuid.New(),
		LBID:       lbID,
		InstanceID: instanceID,
		Port:       port,
		Weight:     weight,
		Health:     "unknown",
	}

	if err := s.lbRepo.AddTarget(ctx, target); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, lb.UserID, "lb.target_add", "loadbalancer", lb.ID.String(), map[string]interface{}{
		"instance_id": instanceID.String(),
		"port":        port,
	})

	return nil
}

func (s *LBService) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	lb, err := s.lbRepo.GetByID(ctx, lbID)
	if err != nil {
		return err
	}

	if err := s.lbRepo.RemoveTarget(ctx, lbID, instanceID); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, lb.UserID, "lb.target_remove", "loadbalancer", lb.ID.String(), map[string]interface{}{
		"instance_id": instanceID.String(),
	})

	return nil
}

func (s *LBService) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	return s.lbRepo.ListTargets(ctx, lbID)
}
