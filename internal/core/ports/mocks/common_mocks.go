package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// VpcRepository is a mock for ports.VpcRepository
type VpcRepository struct {
	mock.Mock
}

func NewVpcRepository(t mock.TestingT) *VpcRepository {
	m := &VpcRepository{}
	m.Test(t)
	return m
}

func (m *VpcRepository) Create(ctx context.Context, vpc *domain.VPC) error {
	args := m.Called(ctx, vpc)
	return args.Error(0)
}

func (m *VpcRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.VPC, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}

func (m *VpcRepository) GetByName(ctx context.Context, name string) (*domain.VPC, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}

func (m *VpcRepository) List(ctx context.Context) ([]*domain.VPC, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.VPC), args.Error(1)
}

func (m *VpcRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// AuditService is a mock for ports.AuditService
type AuditService struct {
	mock.Mock
}

func NewAuditService(t mock.TestingT) *AuditService {
	m := &AuditService{}
	m.Test(t)
	return m
}

func (m *AuditService) Log(ctx context.Context, userID uuid.UUID, action, resourceType, resourceID string, metadata map[string]interface{}) error {
	args := m.Called(ctx, userID, action, resourceType, resourceID, metadata)
	return args.Error(0)
}

func (m *AuditService) ListLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}

// EventService is a mock for ports.EventService
type EventService struct {
	mock.Mock
}

func NewEventService(t mock.TestingT) *EventService {
	m := &EventService{}
	m.Test(t)
	return m
}

func (m *EventService) RecordEvent(ctx context.Context, action, resourceID, resourceType string, metadata map[string]interface{}) error {
	args := m.Called(ctx, action, resourceID, resourceType, metadata)
	return args.Error(0)
}

func (m *EventService) ListEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}
