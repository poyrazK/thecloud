package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockLBRepo is unique to this test file. vpc and instance mocks are reused from dashboard_test.go
type mockLBRepo struct {
	mock.Mock
}

func (m *mockLBRepo) Create(ctx context.Context, lb *domain.LoadBalancer) error {
	args := m.Called(ctx, lb)
	return args.Error(0)
}

func (m *mockLBRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoadBalancer), args.Error(1)
}

func (m *mockLBRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.LoadBalancer, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LoadBalancer), args.Error(1)
}

func (m *mockLBRepo) List(ctx context.Context) ([]*domain.LoadBalancer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LoadBalancer), args.Error(1)
}

func (m *mockLBRepo) ListAll(ctx context.Context) ([]*domain.LoadBalancer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LoadBalancer), args.Error(1)
}

func (m *mockLBRepo) Update(ctx context.Context, lb *domain.LoadBalancer) error {
	args := m.Called(ctx, lb)
	return args.Error(0)
}

func (m *mockLBRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockLBRepo) AddTarget(ctx context.Context, target *domain.LBTarget) error {
	args := m.Called(ctx, target)
	return args.Error(0)
}

func (m *mockLBRepo) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error {
	args := m.Called(ctx, lbID, instanceID)
	return args.Error(0)
}

func (m *mockLBRepo) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	args := m.Called(ctx, lbID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LBTarget), args.Error(1)
}

func (m *mockLBRepo) UpdateTargetHealth(ctx context.Context, lbID, instanceID uuid.UUID, health string) error {
	args := m.Called(ctx, lbID, instanceID, health)
	return args.Error(0)
}

func (m *mockLBRepo) GetTargetsForInstance(ctx context.Context, instanceID uuid.UUID) ([]*domain.LBTarget, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LBTarget), args.Error(1)
}

func TestLBService_Create(t *testing.T) {
	lbRepo := new(mockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instRepo := new(MockInstanceRepo)
	auditSvc := new(services.MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instRepo, auditSvc)

	ctx := context.Background()
	vpcID := uuid.New()
	name := "test-lb"
	port := 80
	algo := "round-robin"

	t.Run("successful creation", func(t *testing.T) {
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
		lbRepo.AssertExpectations(t)
		vpcRepo.AssertExpectations(t)
	})

	t.Run("idempotency check", func(t *testing.T) {
		existing := &domain.LoadBalancer{ID: uuid.New(), Name: name, IdempotencyKey: "key1"}
		lbRepo.On("GetByIdempotencyKey", ctx, "key1").Return(existing, nil).Once()

		lb, err := svc.Create(ctx, name, vpcID, port, algo, "key1")

		assert.NoError(t, err)
		assert.Equal(t, existing.ID, lb.ID)
		lbRepo.AssertExpectations(t)
	})

	t.Run("vpc not found", func(t *testing.T) {
		lbRepo.On("GetByIdempotencyKey", ctx, "key2").Return(nil, errors.New(errors.NotFound, "not found")).Once()
		vpcRepo.On("GetByID", ctx, vpcID).Return(nil, errors.New(errors.NotFound, "not found")).Once()

		lb, err := svc.Create(ctx, name, vpcID, port, algo, "key2")

		assert.Error(t, err)
		assert.Nil(t, lb)
		assert.True(t, errors.Is(err, errors.NotFound))
		vpcRepo.AssertExpectations(t)
	})
}

func TestLBService_PropagatesUserID(t *testing.T) {
	lbRepo := new(mockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instRepo := new(MockInstanceRepo)
	auditSvc := new(services.MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instRepo, auditSvc)

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
	lbRepo.AssertExpectations(t)
}

func TestLBService_AddTarget(t *testing.T) {
	lbRepo := new(mockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instRepo := new(MockInstanceRepo)
	auditSvc := new(services.MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instRepo, auditSvc)

	ctx := context.Background()
	lbID := uuid.New()
	vpcID := uuid.New()
	instID := uuid.New()

	t.Run("successful add target", func(t *testing.T) {
		lbRepo.On("GetByID", ctx, lbID).Return(&domain.LoadBalancer{ID: lbID, VpcID: vpcID}, nil).Once()
		instRepo.On("GetByID", ctx, instID).Return(&domain.Instance{ID: instID, VpcID: &vpcID}, nil).Once()
		lbRepo.On("AddTarget", ctx, mock.Anything).Return(nil).Once()
		auditSvc.On("Log", ctx, mock.Anything, "lb.target_add", "loadbalancer", lbID.String(), mock.Anything).Return(nil).Once()

		err := svc.AddTarget(ctx, lbID, instID, 80, 1)

		assert.NoError(t, err)
		lbRepo.AssertExpectations(t)
		instRepo.AssertExpectations(t)
	})

	t.Run("cross-vpc rejection", func(t *testing.T) {
		otherVpcID := uuid.New()
		lbRepo.On("GetByID", ctx, lbID).Return(&domain.LoadBalancer{ID: lbID, VpcID: vpcID}, nil).Once()
		instRepo.On("GetByID", ctx, instID).Return(&domain.Instance{ID: instID, VpcID: &otherVpcID}, nil).Once()

		err := svc.AddTarget(ctx, lbID, instID, 80, 1)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrLBCrossVPC, err)
		lbRepo.AssertExpectations(t)
		instRepo.AssertExpectations(t)
	})
}

func TestLBService_Delete_Success(t *testing.T) {
	lbRepo := new(mockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instRepo := new(MockInstanceRepo)
	auditSvc := new(services.MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instRepo, auditSvc)

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
	lbRepo.AssertExpectations(t)
}
