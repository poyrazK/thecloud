// Package ports defines service and repository interfaces.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// AutoScalingRepository manages the persistence and metric retrieval for autoscaling groups and policies.
type AutoScalingRepository interface {
	// Scaling Groups
	// CreateGroup saves a new scaling group configuration.
	CreateGroup(ctx context.Context, group *domain.ScalingGroup) error
	// GetGroupByID retrieves a scaling group by its unique ID.
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error)
	// GetGroupByIdempotencyKey retrieves a group using its idempotency key to prevent duplicate creation.
	GetGroupByIdempotencyKey(ctx context.Context, key string) (*domain.ScalingGroup, error)
	// ListGroups returns all scaling groups (often filtered by owner context).
	ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
	// ListAllGroups returns every scaling group in the system (for background runners).
	ListAllGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
	// CountGroupsByVPC counts groups within a specific VPC for limit enforcement.
	CountGroupsByVPC(ctx context.Context, vpcID uuid.UUID) (int, error)
	// UpdateGroup modifies an existing scaling group's configuration or state.
	UpdateGroup(ctx context.Context, group *domain.ScalingGroup) error
	// DeleteGroup removes a scaling group from persistence.
	DeleteGroup(ctx context.Context, id uuid.UUID) error

	// Policies
	// CreatePolicy saves a new scaling policy.
	CreatePolicy(ctx context.Context, policy *domain.ScalingPolicy) error
	// GetPoliciesForGroup retrieves all policies associated with a specific group.
	GetPoliciesForGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.ScalingPolicy, error)
	// GetAllPolicies fetches policies for multiple groups in a single batch.
	GetAllPolicies(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]*domain.ScalingPolicy, error)
	// UpdatePolicyLastScaled updates the timestamp of the last scaling action to enforce cooldowns.
	UpdatePolicyLastScaled(ctx context.Context, policyID uuid.UUID, t time.Time) error
	// DeletePolicy removes a scaling policy.
	DeletePolicy(ctx context.Context, id uuid.UUID) error

	// Group Instances
	// AddInstanceToGroup links a compute instance to a scaling group.
	AddInstanceToGroup(ctx context.Context, groupID, instanceID uuid.UUID) error
	// RemoveInstanceFromGroup unlinks a compute instance from a scaling group.
	RemoveInstanceFromGroup(ctx context.Context, groupID, instanceID uuid.UUID) error
	// GetInstancesInGroup lists the IDs of all instances belonging to a group.
	GetInstancesInGroup(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error)
	// GetAllScalingGroupInstances fetches instances for multiple groups in one batch query to prevent N+1 issues.
	GetAllScalingGroupInstances(ctx context.Context, groupIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)

	// Metrics
	// GetAverageCPU calculates the mean CPU utilization across a set of instances since the given time.
	GetAverageCPU(ctx context.Context, instanceIDs []uuid.UUID, since time.Time) (float64, error)
}

// CreateScalingGroupParams encapsulates arguments for creating a new autoscaling group.
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

// CreateScalingPolicyParams encapsulates arguments for creating a new autoscaling policy.
type CreateScalingPolicyParams struct {
	GroupID     uuid.UUID
	Name        string
	MetricType  string
	TargetValue float64
	ScaleOut    int
	ScaleIn     int
	CooldownSec int
}

// AutoScalingService coordinates the management and enforcement of horizontal scaling rules.
type AutoScalingService interface {
	// CreateGroup establishes a new autoscaling managed set.
	CreateGroup(ctx context.Context, params CreateScalingGroupParams) (*domain.ScalingGroup, error)
	// GetGroup retrieves group details by ID.
	GetGroup(ctx context.Context, id uuid.UUID) (*domain.ScalingGroup, error)
	// ListGroups lists groups authorized for the current caller.
	ListGroups(ctx context.Context) ([]*domain.ScalingGroup, error)
	// DeleteGroup removes a group and its associated policies/instances.
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	// SetDesiredCapacity manually overrides the desired instance count of a group.
	SetDesiredCapacity(ctx context.Context, groupID uuid.UUID, desired int) error

	// CreatePolicy adds a dynamic scaling rule to a group.
	CreatePolicy(ctx context.Context, params CreateScalingPolicyParams) (*domain.ScalingPolicy, error)
	// DeletePolicy removes a specific scaling rule.
	DeletePolicy(ctx context.Context, id uuid.UUID) error
}

// Clock interface allows abstracting wall-clock time for deterministic testing.
type Clock interface {
	Now() time.Time
}

// RealClock provides a concrete implementation of Clock using the system's current time.
type RealClock struct{}

// Now returns the current system time.
func (RealClock) Now() time.Time { return time.Now() }
