package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type MockLBRepo struct {
	LBs map[uuid.UUID]*domain.LoadBalancer
}

func NewMockLBRepo() *MockLBRepo {
	return &MockLBRepo{
		LBs: make(map[uuid.UUID]*domain.LoadBalancer),
	}
}

func (m *MockLBRepo) Create(ctx context.Context, lb *domain.LoadBalancer) error {
	m.LBs[lb.ID] = lb
	return nil
}

func (m *MockLBRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.LoadBalancer, error) {
	if lb, ok := m.LBs[id]; ok {
		return lb, nil
	}
	return nil, nil
}

func (m *MockLBRepo) Update(ctx context.Context, lb *domain.LoadBalancer) error {
	m.LBs[lb.ID] = lb
	return nil
}

func (m *MockLBRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.LBs, id)
	return nil
}

func (m *MockLBRepo) List(ctx context.Context, filter ports.LBFilter) ([]*domain.LoadBalancer, error) {
	var list []*domain.LoadBalancer
	for _, lb := range m.LBs {
		list = append(list, lb)
	}
	return list, nil
}

func (m *MockLBRepo) AddTarget(ctx context.Context, target *domain.LBTarget) error       { return nil }
func (m *MockLBRepo) RemoveTarget(ctx context.Context, lbID, instanceID uuid.UUID) error { return nil }
func (m *MockLBRepo) ListTargets(ctx context.Context, lbID uuid.UUID) ([]*domain.LBTarget, error) {
	return nil, nil
}
func (m *MockLBRepo) UpdateTargetHealth(ctx context.Context, targetID uuid.UUID, health string) error {
	return nil
}
func (m *MockLBRepo) GetListener(ctx context.Context, port int) (*domain.LoadBalancer, error) {
	return nil, nil
}

// Ensure interface satisfaction
var _ ports.LBRepository = (*MockLBRepo)(nil)
