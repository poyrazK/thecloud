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
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStackServiceUnit(t *testing.T) {
	t.Run("CRUD", testStackServiceUnitCRUD)
	t.Run("RBACErrors", testStackServiceUnitRbacErrors)
	t.Run("RepoErrors", testStackServiceUnitRepoErrors)
	t.Run("DeleteErrors", testStackServiceUnitDeleteErrors)
}

func testStackServiceUnitCRUD(t *testing.T) {
	mockRepo := new(MockStackRepo)
	mockInstSvc := new(MockInstanceService)
	mockVpcSvc := new(MockVpcService)
	mockVolSvc := new(MockVolumeService)
	mockSnapSvc := new(MockSnapshotService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewStackService(mockRepo, rbacSvc, mockInstSvc, mockVpcSvc, mockVolSvc, mockSnapSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("CreateStack", func(t *testing.T) {
		done := make(chan struct{})
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Run(func(mock.Arguments) { close(done) }).Maybe()
		mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil).Maybe()

		stack, err := svc.CreateStack(ctx, "my-stack", "Resources: {}", nil)
		require.NoError(t, err)
		assert.NotNil(t, stack)
		assert.Equal(t, "my-stack", stack.Name)

		<-done
	})

	t.Run("GetStack", func(t *testing.T) {
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(&domain.Stack{ID: id}, nil).Once()
		res, err := svc.GetStack(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, id, res.ID)
	})

	t.Run("ListStacks", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("ListByUserID", mock.Anything, userID).Return([]*domain.Stack{{ID: uuid.New()}}, nil).Once()
		res, err := svc.ListStacks(ctx)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("DeleteStack", func(t *testing.T) {
		id := uuid.New()
		stack := &domain.Stack{ID: id, UserID: userID}
		done := make(chan struct{})
		mockRepo.On("GetByID", mock.Anything, id).Return(stack, nil).Once()
		mockRepo.On("ListResources", mock.Anything, id).Return([]domain.StackResource{}, nil).Maybe()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Run(func(mock.Arguments) { close(done) }).Once()

		err := svc.DeleteStack(ctx, id)
		require.NoError(t, err)
		<-done
	})

	t.Run("ValidateTemplate_Valid", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		template := `
Resources:
  MyVPC:
    Type: VPC
    Properties:
      CIDRBlock: 10.0.0.0/16
`
		res, err := svc.ValidateTemplate(ctx, template)
		require.NoError(t, err)
		assert.True(t, res.Valid)
	})

	t.Run("ValidateTemplate_Invalid", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		res, err := svc.ValidateTemplate(ctx, "invalid: yaml: :")
		require.NoError(t, err)
		assert.False(t, res.Valid)
		assert.NotEmpty(t, res.Errors)
	})

	t.Run("ValidateTemplate_Empty", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		res, err := svc.ValidateTemplate(ctx, "Resources: {}")
		require.NoError(t, err)
		assert.False(t, res.Valid)
		assert.Contains(t, res.Errors[0], "at least one resource")
	})
}

func testStackServiceUnitRbacErrors(t *testing.T) {
	mockRepo := new(MockStackRepo)
	mockInstSvc := new(MockInstanceService)
	mockVpcSvc := new(MockVpcService)
	mockVolSvc := new(MockVolumeService)
	mockSnapSvc := new(MockSnapshotService)
	rbacSvc := new(MockRBACService)

	svc := services.NewStackService(mockRepo, rbacSvc, mockInstSvc, mockVpcSvc, mockVolSvc, mockSnapSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	type rbacCase struct {
		name       string
		permission domain.Permission
		resourceID string
		invoke     func(id string) error
	}

	cases := []rbacCase{
		{
			name:       "CreateStack_Unauthorized",
			permission: domain.PermissionStackCreate,
			resourceID: "*",
			invoke: func(id string) error {
				_, err := svc.CreateStack(ctx, "my-stack", "Resources: {}", nil)
				return err
			},
		},
		{
			name:       "GetStack_Unauthorized",
			permission: domain.PermissionStackRead,
			resourceID: uuid.New().String(),
			invoke: func(id string) error {
				_, err := svc.GetStack(ctx, uuid.MustParse(id))
				return err
			},
		},
		{
			name:       "ListStacks_Unauthorized",
			permission: domain.PermissionStackRead,
			resourceID: "*",
			invoke: func(id string) error {
				_, err := svc.ListStacks(ctx)
				return err
			},
		},
		{
			name:       "DeleteStack_Unauthorized",
			permission: domain.PermissionStackDelete,
			resourceID: uuid.New().String(),
			invoke: func(id string) error {
				return svc.DeleteStack(ctx, uuid.MustParse(id))
			},
		},
		{
			name:       "ValidateTemplate_Unauthorized",
			permission: domain.PermissionStackRead,
			resourceID: "*",
			invoke: func(id string) error {
				_, err := svc.ValidateTemplate(ctx, "Resources:\n  Vpc:")
				return err
			},
		},
	}

	authErr := errors.New(errors.Forbidden, "permission denied")
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rbacSvc.On("Authorize", mock.Anything, userID, tenantID, c.permission, c.resourceID).Return(authErr).Once()
			err := c.invoke(c.resourceID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "permission denied")
		})
	}
}

func testStackServiceUnitRepoErrors(t *testing.T) {
	mockRepo := new(MockStackRepo)
	mockInstSvc := new(MockInstanceService)
	mockVpcSvc := new(MockVpcService)
	mockVolSvc := new(MockVolumeService)
	mockSnapSvc := new(MockSnapshotService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewStackService(mockRepo, rbacSvc, mockInstSvc, mockVpcSvc, mockVolSvc, mockSnapSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("CreateStack_RepoError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(fmt.Errorf("db error")).Once()

		_, err := svc.CreateStack(ctx, "my-stack", "Resources: {}", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("GetStack_NotFound", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		_, err := svc.GetStack(ctx, id)
		require.Error(t, err)
	})

	t.Run("GetStack_RepoError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, fmt.Errorf("db error")).Once()

		_, err := svc.GetStack(ctx, id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("ListStacks_RepoError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("ListByUserID", mock.Anything, mock.Anything).Return([]*domain.Stack(nil), fmt.Errorf("db error")).Once()

		_, err := svc.ListStacks(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("DeleteStack_GetByIDError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		err := svc.DeleteStack(ctx, id)
		require.Error(t, err)
	})
}

func testStackServiceUnitDeleteErrors(t *testing.T) {
	mockRepo := new(MockStackRepo)
	mockInstSvc := new(MockInstanceService)
	mockVpcSvc := new(MockVpcService)
	mockVolSvc := new(MockVolumeService)
	mockSnapSvc := new(MockSnapshotService)
	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	svc := services.NewStackService(mockRepo, rbacSvc, mockInstSvc, mockVpcSvc, mockVolSvc, mockSnapSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("DeleteStack_ListResourcesError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		stack := &domain.Stack{ID: id, UserID: userID, TenantID: tenantID}
		done := make(chan struct{})
		mockRepo.On("GetByID", mock.Anything, id).Return(stack, nil).Once()
		mockRepo.On("ListResources", mock.Anything, id).Return(nil, fmt.Errorf("list error")).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Run(func(mock.Arguments) { close(done) }).Once()

		err := svc.DeleteStack(ctx, id)
		require.NoError(t, err)
		<-done
	})

	t.Run("DeleteStack_VPCDeleteError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		stack := &domain.Stack{ID: id, UserID: userID, TenantID: tenantID}
		vpcID := uuid.New()
		done := make(chan struct{})
		mockRepo.On("GetByID", mock.Anything, id).Return(stack, nil).Once()
		mockRepo.On("ListResources", mock.Anything, id).Return([]domain.StackResource{
			{StackID: id, LogicalID: "MyVPC", PhysicalID: vpcID.String(), ResourceType: "VPC"},
		}, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Run(func(mock.Arguments) { close(done) }).Once()
		mockVpcSvc.On("DeleteVPC", mock.Anything, vpcID.String(), true).Return(fmt.Errorf("vpc delete error")).Once()

		err := svc.DeleteStack(ctx, id)
		require.NoError(t, err)
		<-done
	})

	t.Run("DeleteStack_VolumeDeleteError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		stack := &domain.Stack{ID: id, UserID: userID, TenantID: tenantID}
		volID := uuid.New()
		done := make(chan struct{})
		mockRepo.On("GetByID", mock.Anything, id).Return(stack, nil).Once()
		mockRepo.On("ListResources", mock.Anything, id).Return([]domain.StackResource{
			{StackID: id, LogicalID: "MyVol", PhysicalID: volID.String(), ResourceType: "Volume"},
		}, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Run(func(mock.Arguments) { close(done) }).Once()
		mockVolSvc.On("DeleteVolume", mock.Anything, volID.String()).Return(fmt.Errorf("vol delete error")).Once()

		err := svc.DeleteStack(ctx, id)
		require.NoError(t, err)
		<-done
	})

	t.Run("DeleteStack_InstanceTerminateError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		stack := &domain.Stack{ID: id, UserID: userID, TenantID: tenantID}
		instID := uuid.New()
		done := make(chan struct{})
		mockRepo.On("GetByID", mock.Anything, id).Return(stack, nil).Once()
		mockRepo.On("ListResources", mock.Anything, id).Return([]domain.StackResource{
			{StackID: id, LogicalID: "MyInst", PhysicalID: instID.String(), ResourceType: "Instance"},
		}, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Run(func(mock.Arguments) { close(done) }).Once()
		mockInstSvc.On("TerminateInstance", mock.Anything, instID.String()).Return(fmt.Errorf("terminate error")).Once()

		err := svc.DeleteStack(ctx, id)
		require.NoError(t, err)
		<-done
	})

	t.Run("DeleteStack_SnapshotDeleteError", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		id := uuid.New()
		stack := &domain.Stack{ID: id, UserID: userID, TenantID: tenantID}
		snapID := uuid.New()
		done := make(chan struct{})
		mockRepo.On("GetByID", mock.Anything, id).Return(stack, nil).Once()
		mockRepo.On("ListResources", mock.Anything, id).Return([]domain.StackResource{
			{StackID: id, LogicalID: "MySnap", PhysicalID: snapID.String(), ResourceType: "Snapshot"},
		}, nil).Once()
		mockRepo.On("Delete", mock.Anything, id).Return(nil).Run(func(mock.Arguments) { close(done) }).Once()
		mockSnapSvc.On("DeleteSnapshot", mock.Anything, snapID).Return(fmt.Errorf("snap delete error")).Once()

		err := svc.DeleteStack(ctx, id)
		require.NoError(t, err)
		<-done
	})
}
