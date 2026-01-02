package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/poyraz/cloud/internal/core/domain"
)

type AutoScalingRepository interface {
	// Scaling Groups
	CreateGroup(ctx context.Context, group *domain.ScalingGroup) error
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error)
	GetGroupByIdempotencyKey(ctx context.Context, key string) (*domain.ScalingGroup, error)
	ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
	CountGroupsByVPC(ctx context.Context, vpcID uuid.UUID) (int, error)
	UpdateGroup(ctx context.Context, group *domain.ScalingGroup) error
	DeleteGroup(ctx context.Context, id uuid.UUID) error

	// Policies
	CreatePolicy(ctx context.Context, policy *domain.ScalingPolicy) error
	GetPoliciesForGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.ScalingPolicy, error)
	GetAllPolicies(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]*domain.ScalingPolicy, error)
	UpdatePolicyLastScaled(ctx context.Context, policyID uuid.UUID, t time.Time) error
	DeletePolicy(ctx context.Context, id uuid.UUID) error

	// Group Instances
	AddInstanceToGroup(ctx context.Context, groupID, instanceID uuid.UUID) error
	RemoveInstanceFromGroup(ctx context.Context, groupID, instanceID uuid.UUID) error
	GetInstancesInGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
	// GetAllScalingGroupInstances fetches instances for multiple groups in one batch query to prevent N+1
	GetAllScalingGroupInstances(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)

	// Metrics
	GetAverageCPU(ctx context.Context, instanceIDs []uuid.UUID, since time.Time) (float64, error)
}

type AutoScalingService interface {
	CreateGroup(ctx context.Context, name string, vpcID uuid.UUID, image string, ports string, min, max, desired int, lbID *uuid.UUID, idempotencyKey string) (*domain.ScalingGroup, error)
	GetGroup(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error)
	ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	SetDesiredCapacity(ctx context.Context, groupID uuid.UUID, desired int) error

	CreatePolicy(ctx context.Context, groupID uuid.UUID, name, metricType string, targetValue float64, scaleOut, scaleIn, cooldownSec int) (*domain.ScalingPolicy, error)
	DeletePolicy(ctx context.Context, id uuid.UUID) error
}

// Clock interface allows mocking time in tests
type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
