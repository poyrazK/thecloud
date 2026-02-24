package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAutoScalingServiceUnit(t *testing.T) {
	repo := new(MockAutoScalingRepo)
	vpcRepo := new(MockVpcRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewAutoScalingService(repo, vpcRepo, auditSvc)
	ctx := context.Background()

	t.Run("CreateGroup", func(t *testing.T) {
		vpcID := uuid.New()
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID}, nil).Once()
		repo.On("CountGroupsByVPC", mock.Anything, vpcID).Return(0, nil).Once()
		repo.On("CreateGroup", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		group, err := svc.CreateGroup(ctx, ports.CreateScalingGroupParams{
			Name:         "test-group",
			VpcID:        vpcID,
			Image:        "ami-123",
			MinInstances: 1,
			MaxInstances: 5,
			DesiredCount: 2,
		})
		require.NoError(t, err)
		assert.NotNil(t, group)
	})

	t.Run("GetGroup", func(t *testing.T) {
		groupID := uuid.New()
		repo.On("GetGroupByID", mock.Anything, groupID).Return(&domain.ScalingGroup{ID: groupID}, nil).Once()

		group, err := svc.GetGroup(ctx, groupID)
		require.NoError(t, err)
		assert.NotNil(t, group)
		assert.Equal(t, groupID, group.ID)
	})

	t.Run("ListGroups", func(t *testing.T) {
		repo.On("ListGroups", mock.Anything).Return([]*domain.ScalingGroup{{ID: uuid.New()}}, nil).Once()

		groups, err := svc.ListGroups(ctx)
		require.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	t.Run("DeleteGroup", func(t *testing.T) {
		groupID := uuid.New()
		repo.On("GetGroupByID", mock.Anything, groupID).Return(&domain.ScalingGroup{ID: groupID}, nil).Once()
		repo.On("UpdateGroup", mock.Anything, mock.MatchedBy(func(g *domain.ScalingGroup) bool {
			return g.ID == groupID && g.Status == "DELETING"
		})).Return(nil).Once()
		repo.On("DeleteGroup", mock.Anything, groupID).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.DeleteGroup(ctx, groupID)
		require.NoError(t, err)
	})

	t.Run("SetDesiredCapacity", func(t *testing.T) {
		groupID := uuid.New()
		repo.On("GetGroupByID", mock.Anything, groupID).Return(&domain.ScalingGroup{ID: groupID, MinInstances: 1, MaxInstances: 10}, nil).Once()
		repo.On("UpdateGroup", mock.Anything, mock.MatchedBy(func(g *domain.ScalingGroup) bool {
			return g.DesiredCount == 5
		})).Return(nil).Once()

		err := svc.SetDesiredCapacity(ctx, groupID, 5)
		require.NoError(t, err)
	})

	t.Run("CreatePolicy", func(t *testing.T) {
		groupID := uuid.New()
		repo.On("GetGroupByID", mock.Anything, groupID).Return(&domain.ScalingGroup{ID: groupID}, nil).Once()
		repo.On("CreatePolicy", mock.Anything, mock.Anything).Return(nil).Once()

		policy, err := svc.CreatePolicy(ctx, ports.CreateScalingPolicyParams{
			GroupID:     groupID,
			Name:        "cpu-high",
			MetricType:  "cpu",
			TargetValue: 70.0,
			CooldownSec: 60,
		})
		require.NoError(t, err)
		assert.NotNil(t, policy)
	})

	t.Run("DeletePolicy", func(t *testing.T) {
		policyID := uuid.New()
		repo.On("DeletePolicy", mock.Anything, policyID).Return(nil).Once()

		err := svc.DeletePolicy(ctx, policyID)
		require.NoError(t, err)
	})
}
