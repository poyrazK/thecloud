package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLBService_Get(t *testing.T) {
	lbRepo := new(MockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instanceRepo := new(MockInstanceRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instanceRepo, auditSvc)

	lbID := uuid.New()
	lb := &domain.LoadBalancer{ID: lbID, Name: "lb-main"}
	lbRepo.On("GetByID", mock.Anything, lbID).Return(lb, nil).Once()

	res, err := svc.Get(context.Background(), lbID)
	assert.NoError(t, err)
	assert.Equal(t, lbID, res.ID)
	lbRepo.AssertExpectations(t)
}

func TestLBService_List(t *testing.T) {
	lbRepo := new(MockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instanceRepo := new(MockInstanceRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instanceRepo, auditSvc)

	lbRepo.On("List", mock.Anything).Return([]*domain.LoadBalancer{{ID: uuid.New()}}, nil).Once()

	lbs, err := svc.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, lbs, 1)
	lbRepo.AssertExpectations(t)
}

func TestLBService_RemoveTarget(t *testing.T) {
	lbRepo := new(MockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instanceRepo := new(MockInstanceRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instanceRepo, auditSvc)

	lbID := uuid.New()
	instanceID := uuid.New()
	lb := &domain.LoadBalancer{ID: lbID, Name: "lb-main", UserID: uuid.New()}

	lbRepo.On("GetByID", mock.Anything, lbID).Return(lb, nil).Once()
	lbRepo.On("RemoveTarget", mock.Anything, lbID, instanceID).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, lb.UserID, "lb.target_remove", "loadbalancer", lb.ID.String(), mock.Anything).Return(nil).Once()

	err := svc.RemoveTarget(context.Background(), lbID, instanceID)
	assert.NoError(t, err)
	lbRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}

func TestLBService_ListTargets(t *testing.T) {
	lbRepo := new(MockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instanceRepo := new(MockInstanceRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instanceRepo, auditSvc)

	lbID := uuid.New()
	targets := []*domain.LBTarget{{ID: uuid.New(), LBID: lbID}}
	lbRepo.On("ListTargets", mock.Anything, lbID).Return(targets, nil).Once()

	res, err := svc.ListTargets(context.Background(), lbID)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	lbRepo.AssertExpectations(t)
}

func TestLBService_Delete(t *testing.T) {
	lbRepo := new(MockLBRepo)
	vpcRepo := new(MockVpcRepo)
	instanceRepo := new(MockInstanceRepo)
	auditSvc := new(MockAuditService)
	svc := services.NewLBService(lbRepo, vpcRepo, instanceRepo, auditSvc)

	lbID := uuid.New()
	lb := &domain.LoadBalancer{ID: lbID, Name: "lb-main", UserID: uuid.New()}

	lbRepo.On("GetByID", mock.Anything, lbID).Return(lb, nil).Once()
	lbRepo.On("Update", mock.Anything, mock.MatchedBy(func(updated *domain.LoadBalancer) bool {
		return updated.Status == domain.LBStatusDeleted
	})).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, lb.UserID, "lb.delete", "loadbalancer", lb.ID.String(), mock.Anything).Return(nil).Once()

	err := svc.Delete(context.Background(), lbID)
	assert.NoError(t, err)
	lbRepo.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}
