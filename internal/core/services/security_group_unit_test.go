package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSecurityGroupService_Unit(t *testing.T) {
	mockRepo := new(MockSecurityGroupRepo)
	mockVpcRepo := new(MockVpcRepo)
	mockNetwork := new(MockNetworkBackend)
	mockAuditSvc := new(MockAuditService)
	svc := services.NewSecurityGroupService(mockRepo, mockVpcRepo, mockNetwork, mockAuditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	vpcID := uuid.New()

	t.Run("CreateGroup", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "security_group.create", "security_group", mock.Anything, mock.Anything).Return(nil).Once()

		sg, err := svc.CreateGroup(ctx, vpcID, "web-sg", "allow http")
		assert.NoError(t, err)
		assert.NotNil(t, sg)
		assert.Equal(t, "web-sg", sg.Name)
		assert.Len(t, sg.Rules, 2) // Default ARP rules
		mockRepo.AssertExpectations(t)
	})

	t.Run("AddRule", func(t *testing.T) {
		sgID := uuid.New()
		sg := &domain.SecurityGroup{ID: sgID, UserID: userID, VPCID: vpcID}
		rule := domain.SecurityRule{Protocol: "tcp", PortMin: 80, PortMax: 80, Direction: domain.RuleIngress}
		
		mockRepo.On("GetByID", mock.Anything, sgID).Return(sg, nil).Once()
		mockRepo.On("AddRule", mock.Anything, mock.Anything).Return(nil).Once()
		mockVpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID, NetworkID: "net-1"}, nil).Once()
		mockNetwork.On("AddFlowRule", mock.Anything, "net-1", mock.Anything).Return(nil).Once()
		mockAuditSvc.On("Log", mock.Anything, userID, "security_group.add_rule", "security_group", sgID.String(), mock.Anything).Return(nil).Once()

		res, err := svc.AddRule(ctx, sgID, rule)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}
