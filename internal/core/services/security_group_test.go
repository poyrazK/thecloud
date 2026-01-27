package services_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupSecurityGroupServiceTest(_ *testing.T) (*MockSecurityGroupRepo, *MockVpcRepo, *MockNetworkBackend, *MockAuditService, ports.SecurityGroupService) {
	repo := new(MockSecurityGroupRepo)
	vpcRepo := new(MockVpcRepo)
	network := new(MockNetworkBackend)
	auditSvc := new(MockAuditService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := services.NewSecurityGroupService(repo, vpcRepo, network, auditSvc, logger)
	return repo, vpcRepo, network, auditSvc, svc
}

const testSGName = "test-sg"

func TestSecurityGroupServiceCreateGroup(t *testing.T) {
	repo, _, _, auditSvc, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	vpcID := uuid.New()

	repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "security_group.create", "security_group", mock.Anything, mock.Anything).Return(nil)

	sg, err := svc.CreateGroup(ctx, vpcID, testSGName, "desc")

	assert.NoError(t, err)
	assert.NotNil(t, sg)
	assert.Equal(t, testSGName, sg.Name)
	assert.Equal(t, vpcID, sg.VPCID)
}

func TestSecurityGroupServiceGetGroup(t *testing.T) {
	repo, _, _, _, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	sgID := uuid.New()
	vpcID := uuid.New()
	sg := &domain.SecurityGroup{ID: sgID, Name: testSGName}

	// By ID
	repo.On("GetByID", mock.Anything, sgID).Return(sg, nil).Once()
	res, err := svc.GetGroup(ctx, sgID.String(), vpcID)
	assert.NoError(t, err)
	assert.Equal(t, sgID, res.ID)

	// By Name
	repo.On("GetByName", mock.Anything, vpcID, testSGName).Return(sg, nil).Once()
	res, err = svc.GetGroup(ctx, testSGName, vpcID)
	assert.NoError(t, err)
	assert.Equal(t, testSGName, res.Name)
}

func TestSecurityGroupServiceListGroups(t *testing.T) {
	repo, _, _, _, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	vpcID := uuid.New()
	sgs := []*domain.SecurityGroup{{ID: uuid.New(), VPCID: vpcID}}

	repo.On("ListByVPC", mock.Anything, vpcID).Return(sgs, nil)

	res, err := svc.ListGroups(ctx, vpcID)
	assert.NoError(t, err)
	assert.Equal(t, sgs, res)
}

func TestSecurityGroupServiceDeleteGroup(t *testing.T) {
	repo, _, _, auditSvc, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	sgID := uuid.New()

	repo.On("Delete", mock.Anything, sgID).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "security_group.delete", "security_group", sgID.String(), mock.Anything).Return(nil)

	err := svc.DeleteGroup(ctx, sgID)
	assert.NoError(t, err)
}

func TestSecurityGroupServiceAddRule(t *testing.T) {
	repo, vpcRepo, network, auditSvc, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	sgID := uuid.New()
	vpcID := uuid.New()
	sg := &domain.SecurityGroup{ID: sgID, UserID: userID, VPCID: vpcID}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "net1"}
	rule := domain.SecurityRule{Protocol: "tcp", PortMin: 80, PortMax: 80, CIDR: testutil.TestAnyCIDR, Direction: domain.RuleIngress}

	repo.On("GetByID", mock.Anything, sgID).Return(sg, nil)
	repo.On("AddRule", mock.Anything, mock.Anything).Return(nil)
	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	network.On("AddFlowRule", mock.Anything, "net1", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "security_group.add_rule", "security_group", sgID.String(), mock.Anything).Return(nil)

	res, err := svc.AddRule(ctx, sgID, rule)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, sgID, res.GroupID)
}

func TestSecurityGroupServiceAttachToInstance(t *testing.T) {
	repo, vpcRepo, network, auditSvc, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	instID := uuid.New()
	sgID := uuid.New()
	vpcID := uuid.New()
	sg := &domain.SecurityGroup{ID: sgID, VPCID: vpcID}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "net1"}

	repo.On("AddInstanceToGroup", mock.Anything, instID, sgID).Return(nil)
	repo.On("GetByID", mock.Anything, sgID).Return(sg, nil)
	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	network.On("AddFlowRule", mock.Anything, "net1", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "security_group.attach", "instance", instID.String(), mock.Anything).Return(nil)

	err := svc.AttachToInstance(ctx, instID, sgID)
	assert.NoError(t, err)
}

func TestSecurityGroupServiceDetachFromInstance(t *testing.T) {
	repo, vpcRepo, _, auditSvc, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	ctx := context.Background()
	instID := uuid.New()
	sgID := uuid.New()
	vpcID := uuid.New()
	sg := &domain.SecurityGroup{ID: sgID, VPCID: vpcID}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "net1"}

	repo.On("RemoveInstanceFromGroup", mock.Anything, instID, sgID).Return(nil)
	repo.On("GetByID", mock.Anything, sgID).Return(sg, nil)
	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)

	auditSvc.On("Log", mock.Anything, mock.Anything, "security_group.detach", "instance", instID.String(), mock.Anything).Return(nil)

	err := svc.DetachFromInstance(ctx, instID, sgID)
	assert.NoError(t, err)
}

func TestSecurityGroupServiceDetachRemovesFlows(t *testing.T) {
	repo, vpcRepo, network, auditSvc, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)
	defer network.AssertExpectations(t)

	ctx := context.Background()
	instID := uuid.New()
	sgID := uuid.New()
	vpcID := uuid.New()

	sg := &domain.SecurityGroup{
		ID:    sgID,
		VPCID: vpcID,
		Rules: []domain.SecurityRule{{ID: uuid.New(), Protocol: "tcp", PortMin: 80, PortMax: 80}},
	}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "net1"}

	repo.On("RemoveInstanceFromGroup", mock.Anything, instID, sgID).Return(nil)
	repo.On("GetByID", mock.Anything, sgID).Return(sg, nil)
	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	network.On("DeleteFlowRule", mock.Anything, "net1", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, mock.Anything, "security_group.detach", "instance", instID.String(), mock.Anything).Return(nil)

	err := svc.DetachFromInstance(ctx, instID, sgID)
	assert.NoError(t, err)
}

func TestSecurityGroupServiceRemoveRule(t *testing.T) {
	repo, vpcRepo, network, auditSvc, svc := setupSecurityGroupServiceTest(t)
	defer repo.AssertExpectations(t)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	ruleID := uuid.New()
	groupID := uuid.New()
	vpcID := uuid.New()

	rule := &domain.SecurityRule{ID: ruleID, GroupID: groupID}
	sg := &domain.SecurityGroup{ID: groupID, UserID: userID, VPCID: vpcID}
	vpc := &domain.VPC{ID: vpcID, NetworkID: "net1"}

	repo.On("GetRuleByID", mock.Anything, ruleID).Return(rule, nil)
	repo.On("GetByID", mock.Anything, groupID).Return(sg, nil)
	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	network.On("DeleteFlowRule", mock.Anything, "net1", mock.Anything).Return(nil)
	repo.On("DeleteRule", mock.Anything, ruleID).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "security_group.remove_rule", "security_group", groupID.String(), mock.Anything).Return(nil)

	err := svc.RemoveRule(ctx, ruleID)
	assert.NoError(t, err)
}

func TestSecurityGroupServiceErrors(t *testing.T) {
	ctx := context.Background()
	sgID := uuid.New()

	t.Run("Create_RepoError", func(t *testing.T) {
		repo, _, _, _, svc := setupSecurityGroupServiceTest(t)
		repo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
		_, err := svc.CreateGroup(ctx, uuid.New(), "n", "d")
		assert.Error(t, err)
	})

	t.Run("Delete_RepoError", func(t *testing.T) {
		repo, _, _, _, svc := setupSecurityGroupServiceTest(t)
		repo.On("Delete", mock.Anything, sgID).Return(assert.AnError)
		err := svc.DeleteGroup(ctx, sgID)
		assert.Error(t, err)
	})

	t.Run("AddRule_GetGroupError", func(t *testing.T) {
		repo, _, _, _, svc := setupSecurityGroupServiceTest(t)
		repo.On("GetByID", mock.Anything, sgID).Return(nil, assert.AnError)
		_, err := svc.AddRule(ctx, sgID, domain.SecurityRule{})
		assert.Error(t, err)
	})
}

func TestSecurityGroupServiceTranslateToFlow(t *testing.T) {
	repo, vpcRepo, network, auditSvc, svc := setupSecurityGroupServiceTest(t)
	ctx := context.Background()
	sgID := uuid.New()
	vpcID := uuid.New()
	vpc := &domain.VPC{ID: vpcID, NetworkID: "net1"}

	rules := []domain.SecurityRule{
		{Protocol: "udp", PortMin: 53, PortMax: 53, CIDR: testutil.TestGoogleDNSCIDR, Direction: domain.RuleIngress},
		{Protocol: "icmp", CIDR: testutil.TestAnyCIDR, Direction: domain.RuleEgress},
		{Protocol: "tcp", PortMin: 1000, PortMax: 2000, CIDR: testutil.TestTenCIDR, Direction: domain.RuleEgress},
	}

	for _, rule := range rules {
		sgWithRule := &domain.SecurityGroup{ID: sgID, VPCID: vpcID, Rules: []domain.SecurityRule{rule}}
		repo.On("GetByID", mock.Anything, sgID).Return(sgWithRule, nil).Once()
		repo.On("AddRule", mock.Anything, mock.Anything).Return(nil).Once()
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil).Once()
		network.On("AddFlowRule", mock.Anything, "net1", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_, err := svc.AddRule(ctx, sgID, rule)
		assert.NoError(t, err)
	}
}
