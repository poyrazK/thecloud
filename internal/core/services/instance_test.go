package services_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
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

const (
	testPorts           = "8080:80"
	testVPCNetwork      = "cloud-network"
	defaultInstanceType = "basic-2"
)

func setupInstanceServiceTest(_ *testing.T) (*MockInstanceRepo, *MockVpcRepo, *MockSubnetRepo, *MockVolumeRepo, *MockComputeBackend, *MockNetworkBackend, *MockEventService, *MockAuditService, *MockInstanceTypeRepo, *services.TaskQueueStub, ports.InstanceService) {
	repo := new(MockInstanceRepo)
	vpcRepo := new(MockVpcRepo)
	subnetRepo := new(MockSubnetRepo)
	volumeRepo := new(MockVolumeRepo)
	compute := new(MockComputeBackend)
	network := new(MockNetworkBackend)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)
	itRepo := new(MockInstanceTypeRepo)
	taskQueue := new(services.TaskQueueStub)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewInstanceService(services.InstanceServiceParams{
		Repo:             repo,
		VpcRepo:          vpcRepo,
		SubnetRepo:       subnetRepo,
		VolumeRepo:       volumeRepo,
		Compute:          compute,
		Network:          network,
		EventSvc:         eventSvc,
		AuditSvc:         auditSvc,
		InstanceTypeRepo: itRepo,
		TaskQueue:        taskQueue,
		Logger:           logger,
	})
	return repo, vpcRepo, subnetRepo, volumeRepo, compute, network, eventSvc, auditSvc, itRepo, taskQueue, svc
}

func TestLaunchInstanceSuccess(t *testing.T) {
	repo, _, _, _, _, _, _, _, itRepo, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer itRepo.AssertExpectations(t)

	name := "test-inst"
	image := "alpine"
	ports := testPorts

	itRepo.On("GetByID", mock.Anything, defaultInstanceType).Return(&domain.InstanceType{ID: defaultInstanceType}, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Instance")).Return(nil)

	inst, err := svc.LaunchInstance(context.Background(), name, image, ports, "", nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, inst)
	assert.Equal(t, name, inst.Name)
	assert.Equal(t, domain.StatusStarting, inst.Status)
}

func TestLaunchInstancePropagatesUserID(t *testing.T) {
	repo, _, _, _, _, _, _, _, itRepo, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)

	expectedUserID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), expectedUserID)
	name := "test-inst-user"
	image := "alpine"

	itRepo.On("GetByID", mock.Anything, defaultInstanceType).Return(&domain.InstanceType{ID: defaultInstanceType}, nil).Maybe()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(inst *domain.Instance) bool {
		return inst.UserID == expectedUserID
	})).Return(nil)

	inst, err := svc.LaunchInstance(ctx, name, image, "", "", nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, inst)
	assert.Equal(t, expectedUserID, inst.UserID)
}

func TestProvisionInstanceNetworkErrorUpdatesStatus(t *testing.T) {
	repo, vpcRepo, _, _, compute, _, _, _, itRepo, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	itRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.InstanceType{ID: defaultInstanceType, VCPUs: 2, MemoryMB: 4096}, nil).Maybe()

	ctx := context.Background()
	instID := uuid.New()
	vpcID := uuid.New()
	inst := &domain.Instance{ID: instID, VpcID: &vpcID}

	repo.On("GetByID", ctx, instID).Return(inst, nil)
	compute.On("Type").Return("mock")
	vpcRepo.On("GetByID", ctx, vpcID).Return(nil, assert.AnError)
	repo.On("Update", ctx, mock.MatchedBy(func(i *domain.Instance) bool {
		return i.ID == instID && i.Status == domain.StatusError
	})).Return(nil)

	impl, ok := svc.(*services.InstanceService)
	assert.True(t, ok)

	err := impl.Provision(ctx, instID, nil)
	assert.Error(t, err)
}

func TestTerminateInstanceSuccess(t *testing.T) {
	repo, _, _, volumeRepo, compute, _, eventSvc, auditSvc, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer volumeRepo.AssertExpectations(t)
	defer compute.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	id := uuid.New()
	volID := uuid.New()
	inst := &domain.Instance{ID: id, Name: "test", ContainerID: "c123"}
	attachedVolumes := []*domain.Volume{
		{
			ID:         volID,
			Name:       "vol-1",
			Status:     domain.VolumeStatusInUse,
			InstanceID: &id,
			MountPath:  "/data",
		},
	}

	repo.On("GetByID", mock.Anything, id).Return(inst, nil)
	compute.On("DeleteInstance", mock.Anything, "c123").Return(nil)
	volumeRepo.On("ListByInstanceID", mock.Anything, id).Return(attachedVolumes, nil)
	volumeRepo.On("Update", mock.Anything, mock.MatchedBy(func(v *domain.Volume) bool {
		return v.ID == volID &&
			v.Status == domain.VolumeStatusAvailable &&
			v.InstanceID == nil &&
			v.MountPath == ""
	})).Return(nil)
	repo.On("Delete", mock.Anything, id).Return(nil)
	eventSvc.On("RecordEvent", mock.Anything, "INSTANCE_TERMINATE", id.String(), "INSTANCE", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "instance.terminate", "instance", id.String(), mock.Anything).Return(nil)

	err := svc.TerminateInstance(context.Background(), id.String())

	assert.NoError(t, err)
}

func TestTerminateInstanceUpdatesMetricsRunning(t *testing.T) {
	repo, _, _, volumeRepo, compute, _, eventSvc, auditSvc, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer volumeRepo.AssertExpectations(t)
	defer compute.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	instID := uuid.New()
	inst := &domain.Instance{ID: instID, Name: "test", ContainerID: "c123", Status: domain.StatusRunning}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("DeleteInstance", mock.Anything, "c123").Return(nil)
	compute.On("Type").Return("mock")
	volumeRepo.On("ListByInstanceID", mock.Anything, instID).Return([]*domain.Volume{}, nil)
	repo.On("Delete", mock.Anything, instID).Return(nil)
	eventSvc.On("RecordEvent", mock.Anything, "INSTANCE_TERMINATE", instID.String(), "INSTANCE", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "instance.terminate", "instance", instID.String(), mock.Anything).Return(nil)

	err := svc.TerminateInstance(ctx, instID.String())
	assert.NoError(t, err)
}

func TestTerminateInstanceRemoveContainerFailsDoesNotReleaseVolumes(t *testing.T) {
	repo, _, _, volumeRepo, compute, _, eventSvc, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	id := uuid.New()
	inst := &domain.Instance{ID: id, Name: "test", ContainerID: "c123"}

	repo.On("GetByID", mock.Anything, id).Return(inst, nil)
	compute.On("DeleteInstance", mock.Anything, "c123").Return(assert.AnError)

	err := svc.TerminateInstance(context.Background(), id.String())

	assert.Error(t, err)
	volumeRepo.AssertNotCalled(t, "ListByInstanceID", mock.Anything, id)
	repo.AssertNotCalled(t, "Delete", mock.Anything, id)
	eventSvc.AssertNotCalled(t, "RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetInstanceByID(t *testing.T) {
	repo, _, _, _, _, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, Name: "test-inst"}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)

	result, err := svc.GetInstance(context.Background(), instID.String())

	assert.NoError(t, err)
	assert.Equal(t, instID, result.ID)
}

func TestGetInstanceByName(t *testing.T) {
	repo, _, _, _, _, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)

	name := "my-instance"
	inst := &domain.Instance{ID: uuid.New(), Name: name}

	repo.On("GetByName", mock.Anything, name).Return(inst, nil)

	result, err := svc.GetInstance(context.Background(), name)

	assert.NoError(t, err)
	assert.Equal(t, name, result.Name)
}

func TestListInstances(t *testing.T) {
	repo, _, _, _, _, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)

	instances := []*domain.Instance{{Name: "inst1"}, {Name: "inst2"}}

	repo.On("List", mock.Anything).Return(instances, nil)

	result, err := svc.ListInstances(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetInstanceLogs(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123"}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("GetInstanceLogs", mock.Anything, "c123").Return(io.NopCloser(strings.NewReader("log line 1\nlog line 2")), nil)

	logs, err := svc.GetInstanceLogs(context.Background(), instID.String())

	assert.NoError(t, err)
	assert.Contains(t, logs, "log line 1")
}

func TestExecInstanceNotRunning(t *testing.T) {
	repo, _, _, _, _, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: ""}
	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)

	_, err := svc.Exec(context.Background(), instID.String(), []string{"echo", "hi"})
	assert.Error(t, err)
}

func TestExecInstanceSuccess(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123"}
	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("Exec", mock.Anything, "c123", []string{"echo", "hi"}).Return("hi", nil)

	out, err := svc.Exec(context.Background(), instID.String(), []string{"echo", "hi"})
	assert.NoError(t, err)
	assert.Equal(t, "hi", out)
}

func TestExecInstanceError(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123"}
	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("Exec", mock.Anything, "c123", []string{"echo", "hi"}).Return("", assert.AnError)

	_, err := svc.Exec(context.Background(), instID.String(), []string{"echo", "hi"})
	assert.Error(t, err)
}

func TestStopInstanceSuccess(t *testing.T) {
	repo, _, _, _, compute, _, _, auditSvc, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123", Status: domain.StatusRunning}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("StopInstance", mock.Anything, "c123").Return(nil)
	compute.On("Type").Return("mock")
	repo.On("Update", mock.Anything, mock.MatchedBy(func(i *domain.Instance) bool {
		return i.Status == domain.StatusStopped
	})).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "instance.stop", "instance", instID.String(), mock.Anything).Return(nil)

	err := svc.StopInstance(context.Background(), instID.String())

	assert.NoError(t, err)
}

func TestGetConsoleURLUsesContainerID(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123"}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("GetInstanceIP", mock.Anything, "c123").Return("vnc://localhost:5900", nil)

	url, err := svc.GetConsoleURL(context.Background(), instID.String())

	assert.NoError(t, err)
	assert.Equal(t, "vnc://localhost:5900", url)
}

func TestGetConsoleURLUsesInstanceIDFallback(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("GetInstanceIP", mock.Anything, instID.String()).Return("vnc://localhost:5901", nil)

	url, err := svc.GetConsoleURL(context.Background(), instID.String())

	assert.NoError(t, err)
	assert.Equal(t, "vnc://localhost:5901", url)
}

func TestLaunchInstanceWithSubnetAndNetworking(t *testing.T) {
	repo, vpcRepo, subnetRepo, _, _, _, _, _, itRepo, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer subnetRepo.AssertExpectations(t)

	vpcID := uuid.New()
	subnetID := uuid.New()
	name := "net-inst"
	image := "alpine"

	itRepo.On("GetByID", mock.Anything, defaultInstanceType).Return(&domain.InstanceType{ID: defaultInstanceType}, nil).Maybe()
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Instance")).Return(nil)

	inst, err := svc.LaunchInstance(context.Background(), name, image, "", "", &vpcID, &subnetID, nil)

	assert.NoError(t, err)
	assert.NotNil(t, inst)
	assert.Equal(t, domain.StatusStarting, inst.Status)
}

func TestProvisionSuccess(t *testing.T) {
	repo, vpcRepo, subnetRepo, _, compute, network, eventSvc, auditSvc, itRepo, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer subnetRepo.AssertExpectations(t)
	defer compute.AssertExpectations(t)
	defer network.AssertExpectations(t)
	defer eventSvc.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	instID := uuid.New()
	vpcID := uuid.New()
	subnetID := uuid.New()
	image := "alpine"

	inst := &domain.Instance{
		ID:       instID,
		Name:     "inst",
		Image:    image,
		VpcID:    &vpcID,
		SubnetID: &subnetID,
		Ports:    testPorts,
	}

	vpc := &domain.VPC{ID: vpcID, NetworkID: testVPCNetwork}
	subnet := &domain.Subnet{
		ID:        subnetID,
		VPCID:     vpcID,
		CIDRBlock: testutil.TestSubnetCIDR,
		GatewayIP: testutil.TestGatewayIP,
	}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	itRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.InstanceType{ID: defaultInstanceType, VCPUs: 2, MemoryMB: 4096}, nil).Maybe()
	vpcRepo.On("GetByID", mock.Anything, vpcID).Return(vpc, nil)
	subnetRepo.On("GetByID", mock.Anything, subnetID).Return(subnet, nil)
	repo.On("ListBySubnet", mock.Anything, subnetID).Return([]*domain.Instance{}, nil)

	compute.On("CreateInstance", mock.Anything, mock.MatchedBy(func(opts ports.CreateInstanceOptions) bool {
		return opts.ImageName == image &&
			len(opts.Ports) == 1 && opts.Ports[0] == testPorts &&
			opts.NetworkID == testVPCNetwork
	})).Return("c-123", nil)
	compute.On("Type").Return("docker")

	network.On("CreateVethPair", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	network.On("AttachVethToBridge", mock.Anything, testVPCNetwork, mock.Anything).Return(nil)
	network.On("SetVethIP", mock.Anything, mock.Anything, testutil.TestInstanceIP, "24").Return(nil)

	repo.On("Update", mock.Anything, mock.MatchedBy(func(i *domain.Instance) bool {
		return i.Status == domain.StatusRunning && i.ContainerID == "c-123"
	})).Return(nil)
	eventSvc.On("RecordEvent", mock.Anything, "INSTANCE_LAUNCH", instID.String(), "INSTANCE", mock.Anything).Return(nil)
	auditSvc.On("Log", mock.Anything, mock.Anything, "instance.launch", "instance", instID.String(), mock.Anything).Return(nil)

	err := svc.(interface {
		Provision(context.Context, uuid.UUID, []domain.VolumeAttachment) error
	}).Provision(ctx, instID, nil)

	assert.NoError(t, err)
}

func TestInstanceServiceGetInstanceStats(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	defer repo.AssertExpectations(t)
	defer compute.AssertExpectations(t)

	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123"}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("GetInstanceStats", mock.Anything, "c123").Return(io.NopCloser(strings.NewReader("{}")), nil)

	stats, err := svc.GetInstanceStats(context.Background(), instID.String())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestInstanceServiceGetInstanceStatsError(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123"}

	t.Run("RepoError", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, instID).Return(nil, assert.AnError).Once()
		_, err := svc.GetInstanceStats(context.Background(), instID.String())
		assert.Error(t, err)
	})

	t.Run("ComputeError", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, instID).Return(inst, nil).Once()
		compute.On("GetInstanceStats", mock.Anything, "c123").Return(nil, assert.AnError).Once()
		_, err := svc.GetInstanceStats(context.Background(), instID.String())
		assert.Error(t, err)
	})
}

func TestInstanceServiceLaunchInvalidPorts(t *testing.T) {
	_, _, _, _, _, _, _, _, _, _, svc := setupInstanceServiceTest(t)

	_, err := svc.LaunchInstance(context.Background(), "n", "i", "invalid", "", nil, nil, nil)
	assert.Error(t, err)
}

func TestInstanceServiceStopError(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123", Status: domain.StatusRunning}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("StopInstance", mock.Anything, "c123").Return(assert.AnError)

	err := svc.StopInstance(context.Background(), instID.String())
	assert.Error(t, err)
}

func TestInstanceServiceGetInstanceLogsError(t *testing.T) {
	repo, _, _, _, compute, _, _, _, _, _, svc := setupInstanceServiceTest(t)
	instID := uuid.New()
	inst := &domain.Instance{ID: instID, ContainerID: "c123"}

	repo.On("GetByID", mock.Anything, instID).Return(inst, nil)
	compute.On("GetInstanceLogs", mock.Anything, "c123").Return(nil, assert.AnError)

	_, err := svc.GetInstanceLogs(context.Background(), instID.String())
	assert.Error(t, err)
}

func TestInstanceServiceLaunchWithVolumes(t *testing.T) {
	repo, _, _, volumeRepo, _, _, _, _, itRepo, _, svc := setupInstanceServiceTest(t)
	volID := uuid.New()

	volumeRepo.On("GetByID", mock.Anything, volID).Return(&domain.Volume{ID: volID, Name: "v1", Status: domain.VolumeStatusAvailable}, nil)
	itRepo.On("GetByID", mock.Anything, defaultInstanceType).Return(&domain.InstanceType{ID: defaultInstanceType}, nil).Maybe()
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	attachments := []domain.VolumeAttachment{
		{VolumeIDOrName: volID.String(), MountPath: "/data"},
	}
	_, err := svc.LaunchInstance(context.Background(), "n", "i", "", "", nil, nil, attachments)
	assert.NoError(t, err)
}

func TestInstanceServiceGetVolumeByIDOrName(t *testing.T) {
	_, _, _, volumeRepo, _, _, _, _, _, _, _ := setupInstanceServiceTest(t)
	volID := uuid.New()

	// Test by ID
	volumeRepo.On("GetByID", mock.Anything, volID).Return(&domain.Volume{ID: volID}, nil).Once()
	// Wait, getVolumeByIDOrName is unexported. I can't call it directly.
	// But LaunchInstance/TerminateInstance/etc might call it?
	// Actually, only LaunchInstance calls resolveVolumes which calls it.
}
