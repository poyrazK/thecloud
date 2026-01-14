// Package services implements core business workflows.
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

// AutoScalingService manages scaling groups and policies.
type AutoScalingService struct {
	repo     ports.AutoScalingRepository
	vpcRepo  ports.VpcRepository
	auditSvc ports.AuditService
}

// NewAutoScalingService constructs an AutoScalingService with its dependencies.
func NewAutoScalingService(repo ports.AutoScalingRepository, vpcRepo ports.VpcRepository, auditSvc ports.AuditService) *AutoScalingService {
	return &AutoScalingService{
		repo:     repo,
		vpcRepo:  vpcRepo,
		auditSvc: auditSvc,
	}
}

func (s *AutoScalingService) CreateGroup(ctx context.Context, params ports.CreateScalingGroupParams) (*domain.ScalingGroup, error) {
	// Idempotency check
	if params.IdempotencyKey != "" {
		if existing, err := s.repo.GetGroupByIdempotencyKey(ctx, params.IdempotencyKey); err == nil && existing != nil {
			return existing, nil
		}
	}

	// Validation
	if params.MaxInstances > domain.MaxInstancesHardLimit {
		return nil, errors.New(errors.InvalidInput, fmt.Sprintf("max_instances cannot exceed %d", domain.MaxInstancesHardLimit))
	}
	if params.MinInstances < 0 {
		return nil, errors.New(errors.InvalidInput, "min_instances cannot be negative")
	}
	if params.MinInstances > params.MaxInstances {
		return nil, errors.New(errors.InvalidInput, "min_instances cannot be greater than max_instances")
	}
	if params.DesiredCount < params.MinInstances || params.DesiredCount > params.MaxInstances {
		return nil, errors.New(errors.InvalidInput, "desired_count must be between min and max instances")
	}

	// Check VPC exists
	if _, err := s.vpcRepo.GetByID(ctx, params.VpcID); err != nil {
		return nil, err
	}

	// Security: Check VPC group limit
	count, err := s.repo.CountGroupsByVPC(ctx, params.VpcID)
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxScalingGroupsPerVPC {
		return nil, errors.New(errors.ResourceLimitExceeded, fmt.Sprintf("VPC already has %d scaling groups (max: %d)", count, domain.MaxScalingGroupsPerVPC))
	}

	group := &domain.ScalingGroup{
		ID:             uuid.New(),
		UserID:         appcontext.UserIDFromContext(ctx),
		IdempotencyKey: params.IdempotencyKey,
		Name:           params.Name,
		VpcID:          params.VpcID,
		LoadBalancerID: params.LoadBalancerID,
		Image:          params.Image,
		Ports:          params.Ports,
		MinInstances:   params.MinInstances,
		MaxInstances:   params.MaxInstances,
		DesiredCount:   params.DesiredCount,
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

func (s *AutoScalingService) CreatePolicy(ctx context.Context, params ports.CreateScalingPolicyParams) (*domain.ScalingPolicy, error) {
	if _, err := s.repo.GetGroupByID(ctx, params.GroupID); err != nil {
		return nil, err
	}

	if params.CooldownSec < domain.MinCooldownSeconds {
		return nil, errors.New(errors.InvalidInput, fmt.Sprintf("cooldown must be at least %d seconds", domain.MinCooldownSeconds))
	}

	policy := &domain.ScalingPolicy{
		ID:             uuid.New(),
		ScalingGroupID: params.GroupID,
		Name:           params.Name,
		MetricType:     params.MetricType,
		TargetValue:    params.TargetValue,
		ScaleOutStep:   params.ScaleOut,
		ScaleInStep:    params.ScaleIn,
		CooldownSec:    params.CooldownSec,
	}

	if err := s.repo.CreatePolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *AutoScalingService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePolicy(ctx, id)
}
