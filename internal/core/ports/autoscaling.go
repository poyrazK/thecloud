package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

type AutoScalingRepository interface {
	// Scaling Groups
	CreateGroup(ctx context.Context, group *domain.ScalingGroup) error
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error)
	GetGroupByIdempotencyKey(ctx context.Context, key string) (*domain.ScalingGroup, error)
	ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
	ListAllGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
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

// CreateScalingGroupParams encapsulates arguments for creating a scaling group
type CreateScalingGroupParams struct {
	Name           string
	VpcID          uuid.UUID
	Image          string
	Ports          string
	MinInstances   int
	MaxInstances   int
	DesiredCount   int
	LoadBalancerID *uuid.UUID
	IdempotencyKey string
}

// CreateScalingPolicyParams encapsulates arguments for creating a scaling policy
type CreateScalingPolicyParams struct {
	GroupID     uuid.UUID
	Name        string
	MetricType  string
	TargetValue float64
	ScaleOut    int
	ScaleIn     int
	CooldownSec int
}

type AutoScalingService interface {
	CreateGroup(ctx context.Context, params CreateScalingGroupParams) (*domain.ScalingGroup, error)
	GetGroup(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error)
	ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	SetDesiredCapacity(ctx context.Context, groupID uuid.UUID, desired int) error

	CreatePolicy(ctx context.Context, params CreateScalingPolicyParams) (*domain.ScalingPolicy, error)
	DeletePolicy(ctx context.Context, id uuid.UUID) error
}

// Clock interface allows mocking time in tests
type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
