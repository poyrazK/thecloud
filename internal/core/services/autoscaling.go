package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

type AutoScalingService struct {
	repo     ports.AutoScalingRepository
	vpcRepo  ports.VpcRepository
	auditSvc ports.AuditService
}

func NewAutoScalingService(repo ports.AutoScalingRepository, vpcRepo ports.VpcRepository, auditSvc ports.AuditService) *AutoScalingService {
	return &AutoScalingService{
		repo:     repo,
		vpcRepo:  vpcRepo,
		auditSvc: auditSvc,
	}
}

func (s *AutoScalingService) CreateGroup(ctx context.Context, name string, vpcID uuid.UUID, image string, ports string, min, max, desired int, lbID *uuid.UUID, idempotencyKey string) (*domain.ScalingGroup, error) {
	// Idempotency check
	if idempotencyKey != "" {
		if existing, err := s.repo.GetGroupByIdempotencyKey(ctx, idempotencyKey); err == nil && existing != nil {
			return existing, nil
		}
	}

	// Validation
	if max > domain.MaxInstancesHardLimit {
		return nil, errors.New(errors.InvalidInput, fmt.Sprintf("max_instances cannot exceed %d", domain.MaxInstancesHardLimit))
	}
	if min < 0 {
		return nil, errors.New(errors.InvalidInput, "min_instances cannot be negative")
	}
	if min > max {
		return nil, errors.New(errors.InvalidInput, "min_instances cannot be greater than max_instances")
	}
	if desired < min || desired > max {
		return nil, errors.New(errors.InvalidInput, "desired_count must be between min and max instances")
	}

	// Check VPC exists
	if _, err := s.vpcRepo.GetByID(ctx, vpcID); err != nil {
		return nil, err
	}

	// Security: Check VPC group limit
	count, err := s.repo.CountGroupsByVPC(ctx, vpcID)
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxScalingGroupsPerVPC {
		return nil, errors.New(errors.ResourceLimitExceeded, fmt.Sprintf("VPC already has %d scaling groups (max: %d)", count, domain.MaxScalingGroupsPerVPC))
	}

	group := &domain.ScalingGroup{
		ID:             uuid.New(),
		UserID:         appcontext.UserIDFromContext(ctx),
		IdempotencyKey: idempotencyKey,
		Name:           name,
		VpcID:          vpcID,
		LoadBalancerID: lbID,
		Image:          image,
		Ports:          ports,
		MinInstances:   min,
		MaxInstances:   max,
		DesiredCount:   desired,
		CurrentCount:   0, // Worker will spawn these
		Status:         domain.ScalingGroupStatusActive,
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, group.UserID, "asg.group_create", "scaling_group", group.ID.String(), map[string]interface{}{
		"name": group.Name,
	})

	return group, nil
}

func (s *AutoScalingService) GetGroup(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error) {
	return s.repo.GetGroupByID(ctx, id)
}

func (s *AutoScalingService) ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error) {
	return s.repo.ListGroups(ctx)
}

func (s *AutoScalingService) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	group, err := s.repo.GetGroupByID(ctx, id)
	if err != nil {
		return err
	}

	// Mark as DELETING to let worker handle cleanup asynchronously
	group.Status = domain.ScalingGroupStatusDeleting
	group.MinInstances = 0 // Allow desired to be 0
	group.DesiredCount = 0 // Stop scaling out immediately
	if err := s.repo.UpdateGroup(ctx, group); err != nil {
		return err
	}

	_ = s.auditSvc.Log(ctx, group.UserID, "asg.group_delete", "scaling_group", group.ID.String(), map[string]interface{}{
		"name": group.Name,
	})

	return nil
}

// SetDesiredCapacity just updates the DB. Worker reconciles.
func (s *AutoScalingService) SetDesiredCapacity(ctx context.Context, groupID uuid.UUID, desired int) error {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}

	if desired < group.MinInstances || desired > group.MaxInstances {
		return errors.New(errors.InvalidInput, fmt.Sprintf("desired must be between %d and %d", group.MinInstances, group.MaxInstances))
	}

	group.DesiredCount = desired
	return s.repo.UpdateGroup(ctx, group)
}

func (s *AutoScalingService) CreatePolicy(ctx context.Context, groupID uuid.UUID, name, metricType string, targetValue float64, scaleOut, scaleIn, cooldownSec int) (*domain.ScalingPolicy, error) {
	if _, err := s.repo.GetGroupByID(ctx, groupID); err != nil {
		return nil, err
	}

	if cooldownSec < domain.MinCooldownSeconds {
		return nil, errors.New(errors.InvalidInput, fmt.Sprintf("cooldown must be at least %d seconds", domain.MinCooldownSeconds))
	}

	policy := &domain.ScalingPolicy{
		ID:             uuid.New(),
		ScalingGroupID: groupID,
		Name:           name,
		MetricType:     metricType,
		TargetValue:    targetValue,
		ScaleOutStep:   scaleOut,
		ScaleInStep:    scaleIn,
		CooldownSec:    cooldownSec,
	}

	if err := s.repo.CreatePolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *AutoScalingService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePolicy(ctx, id)
}
