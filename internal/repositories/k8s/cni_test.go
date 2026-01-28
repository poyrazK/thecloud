package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInstallCNIWithServiceExecutor(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	masterIP := kubeMasterIP
	cluster := &domain.Cluster{ID: clusterID}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: masterIP}, nil)

	calicoURL := fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/calico.yaml", calicoVersion)
	expectedCmd := []string{kubeShell, "-c", fmt.Sprintf(kubectlApply, calicoURL)}
	instSvc.On("Exec", mock.Anything, instanceID.String(), expectedCmd).Return("ok", nil).Once()

	err := p.installCNI(context.Background(), cluster, masterIP)
	assert.NoError(t, err)
	instSvc.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestPatchKubeProxyWithServiceExecutor(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	masterIP := kubeMasterIP
	cluster := &domain.Cluster{ID: clusterID}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: masterIP}, nil)

	patchCmd := kubectlBase + ` -n kube-system patch configmap kube-proxy --type='json' -p='[{"op": "replace", "path": "/data/config.conf", "value": "apiVersion: kubeproxy.config.k8s.io/v1alpha1\nkind: KubeProxyConfiguration\nmode: \"\"\nconntrack:\n  maxPerCore: 0"}]'`
	restartCmd := kubectlBase + " -n kube-system delete pod -l k8s-app=kube-proxy"

	instSvc.On("Exec", mock.Anything, instanceID.String(), []string{kubeShell, "-c", patchCmd}).Return("ok", nil).Once()
	instSvc.On("Exec", mock.Anything, instanceID.String(), []string{kubeShell, "-c", restartCmd}).Return("ok", nil).Once()

	err := p.patchKubeProxy(context.Background(), cluster, masterIP)
	assert.NoError(t, err)
	instSvc.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestInstallObservabilityWithServiceExecutor(t *testing.T) {
	instSvc := new(mockInstanceService)
	repo := new(mockClusterRepo)
	p := newProvisioner(instSvc, repo)

	clusterID := uuid.New()
	instanceID := uuid.New()
	masterIP := kubeMasterIP
	cluster := &domain.Cluster{ID: clusterID}

	repo.On("GetNodes", mock.Anything, clusterID).Return([]*domain.ClusterNode{{ID: uuid.New(), InstanceID: instanceID}}, nil)
	instSvc.On("GetInstance", mock.Anything, instanceID.String()).Return(&domain.Instance{ID: instanceID, PrivateIP: masterIP}, nil)

	ksmManifest := "https://github.com/kubernetes/kube-state-metrics/releases/download/v2.10.1/standard.yaml"
	expectedCmd := []string{kubeShell, "-c", fmt.Sprintf(kubectlApply, ksmManifest)}
	instSvc.On("Exec", mock.Anything, instanceID.String(), expectedCmd).Return("ok", nil).Once()

	err := p.installObservability(context.Background(), cluster, masterIP)
	assert.NoError(t, err)
	instSvc.AssertExpectations(t)
	repo.AssertExpectations(t)
}
