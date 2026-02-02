package mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

type MockAuditService struct{}

func NewMockAuditService() ports.AuditService {
	return &MockAuditService{}
}

func (m *MockAuditService) Log(ctx context.Context, action, resourceType string, resourceID uuid.UUID, metadata map[string]interface{}) error {
	return nil
}

func (m *MockAuditService) InitialLog(ctx context.Context, action domain.AuditAction, resource domain.AuditResource, resourceID uuid.UUID, metadata map[string]interface{}) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *MockAuditService) CompleteLog(ctx context.Context, eventID uuid.UUID, status domain.AuditStatus, err error) error {
	return nil
}

func (m *MockAuditService) LogEvent(ctx context.Context, event *domain.Event) error {
	return nil
}

func (m *MockAuditService) ListLogs(ctx context.Context, filter ports.AuditFilter) ([]*domain.AuditLog, error) {
	return nil, nil
}
