package services_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSubnetService_Unit(t *testing.T) {
	repo := new(MockSubnetRepo)
	vpcRepo := new(MockVpcRepo)
	auditSvc := new(MockAuditService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewSubnetService(services.SubnetServiceParams{
		Repo:     repo,
		RBACSvc:  rbacSvc,
		VpcRepo:  vpcRepo,
		AuditSvc: auditSvc,
		Logger:   slog.Default(),
	})

	ctx := context.Background()
	userID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("CreateSubnet", func(t *testing.T) {
		vpcID := uuid.New()
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID, CIDRBlock: "10.0.0.0/16"}, nil).Once()
		repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "subnet.create", "subnet", mock.Anything, mock.Anything).Return(nil).Once()

		subnet, err := svc.CreateSubnet(ctx, vpcID, "test-subnet", "10.0.1.0/24", "us-east-1a")
		require.NoError(t, err)
		assert.NotNil(t, subnet)
		assert.Equal(t, "10.0.1.1", subnet.GatewayIP)
		repo.AssertExpectations(t)
	})

	t.Run("CreateSubnet_VPCNotFound", func(t *testing.T) {
		vpcID := uuid.New()
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(nil, fmt.Errorf("not found")).Once()

		_, err := svc.CreateSubnet(ctx, vpcID, "fail", "10.0.1.0/24", "az1")
		require.Error(t, err)
	})

	t.Run("CreateSubnet_InvalidCIDR", func(t *testing.T) {
		vpcID := uuid.New()
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID, CIDRBlock: "10.0.0.0/16"}, nil).Once()

		_, err := svc.CreateSubnet(ctx, vpcID, "fail", "invalid", "az1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid subnet CIDR")
	})

	t.Run("CreateSubnet_OutOfRange", func(t *testing.T) {
		vpcID := uuid.New()
		vpcRepo.On("GetByID", mock.Anything, vpcID).Return(&domain.VPC{ID: vpcID, CIDRBlock: "10.0.0.0/16"}, nil).Once()

		_, err := svc.CreateSubnet(ctx, vpcID, "fail", "192.168.1.0/24", "az1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "within VPC CIDR range")
	})

	t.Run("GetSubnet", func(t *testing.T) {
		id := uuid.New()
		vpcID := uuid.New()

		t.Run("ByID", func(t *testing.T) {
			repo.On("GetByID", mock.Anything, id).Return(&domain.Subnet{ID: id}, nil).Once()
			res, err := svc.GetSubnet(ctx, id.String(), vpcID)
			require.NoError(t, err)
			assert.Equal(t, id, res.ID)
		})

		t.Run("ByName", func(t *testing.T) {
			repo.On("GetByName", mock.Anything, vpcID, "subnet-1").Return(&domain.Subnet{Name: "subnet-1"}, nil).Once()
			res, err := svc.GetSubnet(ctx, "subnet-1", vpcID)
			require.NoError(t, err)
			assert.Equal(t, "subnet-1", res.Name)
		})
	})

	t.Run("DeleteSubnet", func(t *testing.T) {
		id := uuid.New()
		repo.On("GetByID", mock.Anything, id).Return(&domain.Subnet{ID: id, UserID: userID}, nil).Once()
		repo.On("Delete", mock.Anything, id).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "subnet.delete", "subnet", id.String(), mock.Anything).Return(nil).Once()

		err := svc.DeleteSubnet(ctx, id)
		require.NoError(t, err)
	})
}
