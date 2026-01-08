package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupLBServiceTest(t *testing.T) (*MockLBRepo, *MockVpcRepo, *MockInstanceRepo, *MockAuditService, ports.LBService) {
	lbRepo := new(MockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instRepo := new(MockInstanceRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instRepo, auditSvc)
	return lbRepo, vpcRepo, instRepo, auditSvc, svc
}

func TestLBService_Create(t *testing.T) {
	ctx := context.Background()
	vpcID := uuid.New()
	name := "test-lb"
	port := 80
	algo := "round-robin"

	t.Run("successful creation", func(t *testing.T) {
		lbRepo, vpcRepo, _, auditSvc, svc := setupLBServiceTest(t)
		defer lbRepo.AssertExpectations(t)
		defer vpcRepo.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		lbRepo.On("GetByIdempotencyKey", ctx, "key1").Return(nil, errors.New(errors.NotFound, "not found")).Once()
		vpcRepo.On("GetByID", ctx, vpcID).Return(&domain.VPC{ID: vpcID}, nil).Once()
		lbRepo.On("Create", ctx, mock.MatchedBy(func(lb *domain.LoadBalancer) bool {
			return lb.Name == name && lb.VpcID == vpcID && lb.Port == port && lb.Status == domain.LBStatusCreating
		})).Return(nil).Once()
		auditSvc.On("Log", ctx, mock.Anything, "lb.create", "loadbalancer", mock.Anything, mock.Anything).Return(nil).Once()

		lb, err := svc.Create(ctx, name, vpcID, port, algo, "key1")

		assert.NoError(t, err)
		assert.NotNil(t, lb)
		assert.Equal(t, name, lb.Name)
		assert.Equal(t, domain.LBStatusCreating, lb.Status)
	})

	t.Run("idempotency check", func(t *testing.T) {
		lbRepo, _, _, _, svc := setupLBServiceTest(t)
		defer lbRepo.AssertExpectations(t)

		existing := &domain.LoadBalancer{ID: uuid.New(), Name: name, IdempotencyKey: "key1"}
		lbRepo.On("GetByIdempotencyKey", ctx, "key1").Return(existing, nil).Once()

		lb, err := svc.Create(ctx, name, vpcID, port, algo, "key1")

		assert.NoError(t, err)
		assert.Equal(t, existing.ID, lb.ID)
	})

	t.Run("vpc not found", func(t *testing.T) {
		lbRepo, vpcRepo, _, _, svc := setupLBServiceTest(t)
		defer lbRepo.AssertExpectations(t)
		defer vpcRepo.AssertExpectations(t)

		lbRepo.On("GetByIdempotencyKey", ctx, "key2").Return(nil, errors.New(errors.NotFound, "not found")).Once()
		vpcRepo.On("GetByID", ctx, vpcID).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		lb, err := svc.Create(ctx, name, vpcID, port, algo, "key2")

		assert.Error(t, err)
		assert.Nil(t, lb)
		assert.True(t, errors.Is(err, errors.NotFound))
	})
}

func TestLBService_PropagatesUserID(t *testing.T) {
	lbRepo, vpcRepo, _, auditSvc, svc := setupLBServiceTest(t)
	defer lbRepo.AssertExpectations(t)
	defer vpcRepo.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	expectedUserID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), expectedUserID)
	vpcID := uuid.New()
	name := "test-lb-user"

	lbRepo.On("GetByIdempotencyKey", ctx, "key3").Return(nil, errors.New(errors.NotFound, "not found")).Once()
	vpcRepo.On("GetByID", ctx, vpcID).Return(&domain.VPC{ID: vpcID}, nil).Once()
	lbRepo.On("Create", ctx, mock.MatchedBy(func(lb *domain.LoadBalancer) bool {
		return lb.UserID == expectedUserID
	})).Return(nil).Once()
	auditSvc.On("Log", ctx, expectedUserID, "lb.create", "loadbalancer", mock.Anything, mock.Anything).Return(nil).Once()

	lb, err := svc.Create(ctx, name, vpcID, 80, "round-robin", "key3")

	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, lb.UserID)
}

func TestLBService_AddTarget(t *testing.T) {
	ctx := context.Background()
	lbID := uuid.New()
	vpcID := uuid.New()
	instID := uuid.New()

	t.Run("successful add target", func(t *testing.T) {
		lbRepo, _, instRepo, auditSvc, svc := setupLBServiceTest(t)
		defer lbRepo.AssertExpectations(t)
		defer instRepo.AssertExpectations(t)
		defer auditSvc.AssertExpectations(t)

		lbRepo.On("GetByID", ctx, lbID).Return(&domain.LoadBalancer{ID: lbID, VpcID: vpcID}, nil).Once()
		instRepo.On("GetByID", ctx, instID).Return(&domain.Instance{ID: instID, VpcID: &vpcID}, nil).Once()
		lbRepo.On("AddTarget", ctx, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", ctx, mock.Anything, "lb.target_add", "loadbalancer", lbID.String(), mock.Anything).Return(nil).Once()

		err := svc.AddTarget(ctx, lbID, instID, 80, 1)

		assert.NoError(t, err)
	})

	t.Run("cross-vpc rejection", func(t *testing.T) {
		lbRepo, _, instRepo, _, svc := setupLBServiceTest(t)
		defer lbRepo.AssertExpectations(t)
		defer instRepo.AssertExpectations(t)

		otherVpcID := uuid.New()
		lbRepo.On("GetByID", ctx, lbID).Return(&domain.LoadBalancer{ID: lbID, VpcID: vpcID}, nil).Once()
		instRepo.On("GetByID", ctx, instID).Return(&domain.Instance{ID: instID, VpcID: &otherVpcID}, nil).Once()

		err := svc.AddTarget(ctx, lbID, instID, 80, 1)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrLBCrossVPC, err)
	})
}

func TestLBService_Delete_Success(t *testing.T) {
	lbRepo, _, _, auditSvc, svc := setupLBServiceTest(t)
	defer lbRepo.AssertExpectations(t)
	defer auditSvc.AssertExpectations(t)

	ctx := context.Background()
	lbID := uuid.New()

	// Expect Get then Update (Soft Delete)
	lb := &domain.LoadBalancer{ID: lbID, Status: domain.LBStatusActive}
	lbRepo.On("GetByID", ctx, lbID).Return(lb, nil)
	lbRepo.On("Update", ctx, mock.MatchedBy(func(l *domain.LoadBalancer) bool {
		return l.ID == lbID && l.Status == domain.LBStatusDeleted
	})).Return(nil)
	auditSvc.On("Log", ctx, mock.Anything, "lb.delete", "loadbalancer", lbID.String(), mock.Anything).Return(nil).Once()

	err := svc.Delete(ctx, lbID)

	assert.NoError(t, err)
}
