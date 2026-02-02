package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type MockGlobalLBRepo struct {
	GLBs      map[uuid.UUID]*domain.GlobalLoadBalancer
	Endpoints map[uuid.UUID][]*domain.GlobalEndpoint
}

func NewMockGlobalLBRepo() *MockGlobalLBRepo {
	return &MockGlobalLBRepo{
		GLBs:      make(map[uuid.UUID]*domain.GlobalLoadBalancer),
		Endpoints: make(map[uuid.UUID][]*domain.GlobalEndpoint),
	}
}

func (m *MockGlobalLBRepo) Create(ctx context.Context, glb *domain.GlobalLoadBalancer) error {
	m.GLBs[glb.ID] = glb
	return nil
}

func (m *MockGlobalLBRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.GlobalLoadBalancer, error) {
	if glb, ok := m.GLBs[id]; ok {
		// return copy
		return glb, nil
	}
	return nil, nil // simplified
}

func (m *MockGlobalLBRepo) GetByHostname(ctx context.Context, hostname string) (*domain.GlobalLoadBalancer, error) {
	for _, glb := range m.GLBs {
		if glb.Hostname == hostname {
			return glb, nil
		}
	}
	return nil, nil
}

func (m *MockGlobalLBRepo) List(ctx context.Context) ([]*domain.GlobalLoadBalancer, error) {
	var list []*domain.GlobalLoadBalancer
	for _, glb := range m.GLBs {
		list = append(list, glb)
	}
	return list, nil
}

func (m *MockGlobalLBRepo) Update(ctx context.Context, glb *domain.GlobalLoadBalancer) error {
	m.GLBs[glb.ID] = glb
	return nil
}

func (m *MockGlobalLBRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.GLBs, id)
	return nil
}

func (m *MockGlobalLBRepo) AddEndpoint(ctx context.Context, ep *domain.GlobalEndpoint) error {
	m.Endpoints[ep.GlobalLBID] = append(m.Endpoints[ep.GlobalLBID], ep)
	return nil
}

func (m *MockGlobalLBRepo) RemoveEndpoint(ctx context.Context, endpointID uuid.UUID) error {
	// inefficient but mock
	for glbID, eps := range m.Endpoints {
		var newEps []*domain.GlobalEndpoint
		for _, ep := range eps {
			if ep.ID != endpointID {
				newEps = append(newEps, ep)
			}
		}
		m.Endpoints[glbID] = newEps
	}
	return nil
}

func (m *MockGlobalLBRepo) ListEndpoints(ctx context.Context, glbID uuid.UUID) ([]*domain.GlobalEndpoint, error) {
	return m.Endpoints[glbID], nil
}

func (m *MockGlobalLBRepo) UpdateEndpointHealth(ctx context.Context, epID uuid.UUID, healthy bool) error {
	for _, eps := range m.Endpoints {
		for _, ep := range eps {
			if ep.ID == epID {
				ep.Healthy = healthy
				return nil
			}
		}
	}
	return nil
}

// Ensure interface satisfaction
var _ ports.GlobalLBRepository = (*MockGlobalLBRepo)(nil)
