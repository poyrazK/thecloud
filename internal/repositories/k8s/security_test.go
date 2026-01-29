package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnsureClusterSecurityGroupExists(t *testing.T) {
	sgSvc := new(MockSecurityGroupService)
	cluster := &domain.Cluster{Name: "dev", VpcID: uuid.New()}
	group := &domain.SecurityGroup{ID: uuid.New()}

	sgSvc.On("GetGroup", mock.Anything, "sg-"+cluster.Name, cluster.VpcID).Return(group, nil).Once()

	p := &KubeadmProvisioner{sgSvc: sgSvc}
	err := p.ensureClusterSecurityGroup(context.Background(), cluster)
	assert.NoError(t, err)

	sgSvc.AssertNotCalled(t, "CreateGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestEnsureClusterSecurityGroupCreatesRules(t *testing.T) {
	sgSvc := new(MockSecurityGroupService)
	cluster := &domain.Cluster{Name: "prod", VpcID: uuid.New()}
	group := &domain.SecurityGroup{ID: uuid.New()}

	sgSvc.On("GetGroup", mock.Anything, "sg-"+cluster.Name, cluster.VpcID).Return(nil, errors.New("not found")).Once()
	sgSvc.On("CreateGroup", mock.Anything, cluster.VpcID, "sg-"+cluster.Name, "Kubernetes cluster security group").Return(group, nil).Once()
	sgSvc.On("AddRule", mock.Anything, group.ID, mock.Anything).Return(&domain.SecurityRule{}, nil).Times(6)

	p := &KubeadmProvisioner{sgSvc: sgSvc}
	err := p.ensureClusterSecurityGroup(context.Background(), cluster)
	assert.NoError(t, err)

	sgSvc.AssertExpectations(t)
}

func TestApplyBaseSecurityNoIsolation(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, NetworkIsolation: false}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: kubeMasterIP}, nil)

	err := p.applyBaseSecurity(context.Background(), cluster, kubeMasterIP)
	assert.NoError(t, err)

	instSvc.AssertNotCalled(t, "Exec", mock.Anything, mock.Anything, mock.Anything)
}

func TestApplyBaseSecurityWithIsolation(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	cluster := &domain.Cluster{ID: clusterID, NetworkIsolation: true}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: kubeMasterIP}, nil)

	instSvc.On("Exec", mock.Anything, instanceID.String(), mock.MatchedBy(func(cmd []string) bool {
		return isShellCommand(cmd, "cat <<EOF > /tmp/base-security.yaml")
	})).Return("ok", nil).Once()
	instSvc.On("Exec", mock.Anything, instanceID.String(), mock.MatchedBy(func(cmd []string) bool {
		return isShellCommand(cmd, kubectlBase+" apply -f /tmp/base-security.yaml")
	})).Return("ok", nil).Once()

	err := p.applyBaseSecurity(context.Background(), cluster, kubeMasterIP)
	assert.NoError(t, err)
	instSvc.AssertExpectations(t)
}

func isShellCommand(cmd []string, contains string) bool {
	if len(cmd) != 3 {
		return false
	}
	if cmd[0] != kubeShell || cmd[1] != "-c" {
		return false
	}
	return strings.Contains(cmd[2], contains)
}
