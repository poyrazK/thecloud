package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuditServiceLog(t *testing.T) {
	repo := new(MockAuditRepo)
	svc := services.NewAuditService(repo)

	userID := uuid.New()
	action := "test.action"
	resType := "instance"
	resID := "123"
	details := map[string]interface{}{"key": "value"}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(log *domain.AuditLog) bool {
		return log.UserID == userID && log.Action == action && log.ResourceType == resType && log.ResourceID == resID
	})).Return(nil)

	err := svc.Log(context.Background(), userID, action, resType, resID, details)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuditServiceListLogs(t *testing.T) {
	repo := new(MockAuditRepo)
	svc := services.NewAuditService(repo)

	userID := uuid.New()
	expectedLogs := []*domain.AuditLog{
		{ID: uuid.New(), UserID: userID, Action: "login", CreatedAt: time.Now()},
	}

	repo.On("ListByUserID", mock.Anything, userID, 50).Return(expectedLogs, nil)

	logs, err := svc.ListLogs(context.Background(), userID, 0) // Test default limit
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, expectedLogs[0].ID, logs[0].ID)
	repo.AssertExpectations(t)
}
