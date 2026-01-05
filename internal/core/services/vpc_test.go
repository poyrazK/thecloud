package services_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVpcService_Create_Success(t *testing.T) {
	vpcRepo := new(MockVpcRepo)
	docker := new(MockDockerClient)
	auditSvc := new(services.MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVpcService(vpcRepo, docker, auditSvc, logger)
	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	name := "test-vpc"

	docker.On("CreateNetwork", ctx, mock.MatchedBy(func(n string) bool {
		return len(n) > 0 // Dynamic name
	})).Return("docker-net-123", nil)
	vpcRepo.On("Create", ctx, mock.MatchedBy(func(vpc *domain.VPC) bool {
		return vpc.Name == name
	})).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "vpc.create", "vpc", mock.Anything, mock.Anything).Return(nil)

	vpc, err := svc.CreateVPC(ctx, name)

	assert.NoError(t, err)
	assert.NotNil(t, vpc)
	assert.Equal(t, name, vpc.Name)
	assert.Equal(t, "docker-net-123", vpc.NetworkID)

	vpcRepo.AssertExpectations(t)
	docker.AssertExpectations(t)
}

func TestVpcService_Create_DBFailure_RollsBackNetwork(t *testing.T) {
	vpcRepo := new(MockVpcRepo)
	docker := new(MockDockerClient)
	auditSvc := new(services.MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVpcService(vpcRepo, docker, auditSvc, logger)
	ctx := context.Background()
	name := "fail-vpc"

	docker.On("CreateNetwork", ctx, mock.Anything).Return("docker-net-456", nil)
	vpcRepo.On("Create", ctx, mock.Anything).Return(assert.AnError)
	docker.On("RemoveNetwork", ctx, "docker-net-456").Return(nil) // Rollback

	vpc, err := svc.CreateVPC(ctx, name)

	assert.Error(t, err)
	assert.Nil(t, vpc)
	docker.AssertCalled(t, "RemoveNetwork", ctx, "docker-net-456")
}

func TestVpcService_Delete_Success(t *testing.T) {
	vpcRepo := new(MockVpcRepo)
	docker := new(MockDockerClient)
	auditSvc := new(services.MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVpcService(vpcRepo, docker, auditSvc, logger)
	ctx := context.Background()
	vpcID := uuid.New()
	vpc := &domain.VPC{
		ID:        vpcID,
		Name:      "to-delete",
		NetworkID: "docker-net-789",
	}

	vpcRepo.On("GetByID", ctx, vpcID).Return(vpc, nil)
	docker.On("RemoveNetwork", ctx, "docker-net-789").Return(nil)
	vpcRepo.On("Delete", ctx, vpcID).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "vpc.delete", "vpc", mock.Anything, mock.Anything).Return(nil)

	err := svc.DeleteVPC(ctx, vpcID.String())

	assert.NoError(t, err)
	vpcRepo.AssertExpectations(t)
	docker.AssertExpectations(t)
}

func TestVpcService_List_Success(t *testing.T) {
	vpcRepo := new(MockVpcRepo)
	docker := new(MockDockerClient)
	auditSvc := new(services.MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVpcService(vpcRepo, docker, auditSvc, logger)
	ctx := context.Background()

	vpcs := []*domain.VPC{{Name: "vpc1"}, {Name: "vpc2"}}
	vpcRepo.On("List", ctx).Return(vpcs, nil)

	result, err := svc.ListVPCs(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	vpcRepo.AssertExpectations(t)
}

func TestVpcService_Get_ByName(t *testing.T) {
	vpcRepo := new(MockVpcRepo)
	docker := new(MockDockerClient)
	auditSvc := new(services.MockAuditService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewVpcService(vpcRepo, docker, auditSvc, logger)
	ctx := context.Background()
	name := "my-vpc"
	vpc := &domain.VPC{ID: uuid.New(), Name: name}

	vpcRepo.On("GetByName", ctx, name).Return(vpc, nil)

	result, err := svc.GetVPC(ctx, name)

	assert.NoError(t, err)
	assert.Equal(t, name, result.Name)
	vpcRepo.AssertExpectations(t)
}
