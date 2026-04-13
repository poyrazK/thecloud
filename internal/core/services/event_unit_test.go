package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	domain "github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEventService_Unit(t *testing.T) {
	mockRepo := new(MockEventRepo)
	mockPublisher := new(MockRealtimePublisher)
	mockPublisher.On("PublishEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	ctx = appcontext.WithTenantID(ctx, tenantID)
	ctx = appcontext.WithUserID(ctx, userID)

	t.Run("RecordEvent_Success", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewEventService(services.EventServiceParams{
			Repo:      mockRepo,
			RBACSvc:   mockRBAC,
			Publisher: mockPublisher,
			Logger:    slog.Default(),
		})
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.RecordEvent(ctx, "TEST_ACTION", "res-123", "TEST", map[string]interface{}{"key": "value"})
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RecordEvent_RepoError", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewEventService(services.EventServiceParams{
			Repo:      mockRepo,
			RBACSvc:   mockRBAC,
			Publisher: mockPublisher,
			Logger:    slog.Default(),
		})
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New(errors.Internal, "db error")).Once()

		err := svc.RecordEvent(ctx, "TEST_ACTION", "res-123", "TEST", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("RecordEvent_PublisherNil_Skips", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewEventService(services.EventServiceParams{
			Repo:    mockRepo,
			RBACSvc: mockRBAC,
			Logger:  slog.Default(),
		})
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.RecordEvent(ctx, "TEST_ACTION", "res-123", "TEST", nil)
		require.NoError(t, err)
	})

	t.Run("RecordEvent_PublisherError_LogsWarning", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewEventService(services.EventServiceParams{
			Repo:      mockRepo,
			RBACSvc:   mockRBAC,
			Publisher: mockPublisher,
			Logger:    slog.Default(),
		})
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		mockPublisher.On("PublishEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(errors.Internal, "redis error")).Once()

		err := svc.RecordEvent(ctx, "TEST_ACTION", "res-123", "TEST", nil)
		require.NoError(t, err) // publisher errors don't fail the method
	})

	t.Run("ListEvents_Success", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewEventService(services.EventServiceParams{
			Repo:      mockRepo,
			RBACSvc:   mockRBAC,
			Publisher: mockPublisher,
			Logger:    slog.Default(),
		})
		events := []*domain.Event{
			{ID: uuid.New(), Action: "A1"},
			{ID: uuid.New(), Action: "A2"},
		}
		mockRepo.On("List", mock.Anything, 10).Return(events, nil).Once()

		result, err := svc.ListEvents(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("ListEvents_DefaultLimit", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewEventService(services.EventServiceParams{
			Repo:      mockRepo,
			RBACSvc:   mockRBAC,
			Publisher: mockPublisher,
			Logger:    slog.Default(),
		})
		mockRepo.On("List", mock.Anything, 50).Return([]*domain.Event{}, nil).Once()

		result, err := svc.ListEvents(ctx, 0)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("ListEvents_RepoError", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		svc := services.NewEventService(services.EventServiceParams{
			Repo:      mockRepo,
			RBACSvc:   mockRBAC,
			Publisher: mockPublisher,
			Logger:    slog.Default(),
		})
		mockRepo.On("List", mock.Anything, 10).Return(nil, errors.New(errors.Internal, "db error")).Once()

		_, err := svc.ListEvents(ctx, 10)
		require.Error(t, err)
	})

	t.Run("ListEvents_Unauthorized", func(t *testing.T) {
		mockRBAC := new(MockRBACService)
		mockRBAC.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(errors.Forbidden, "permission denied"))
		svc := services.NewEventService(services.EventServiceParams{
			Repo:      mockRepo,
			RBACSvc:   mockRBAC,
			Publisher: mockPublisher,
			Logger:    slog.Default(),
		})

		_, err := svc.ListEvents(ctx, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})
}

func TestNewEventService(t *testing.T) {
	t.Run("NilLogger_UsesDefault", func(t *testing.T) {
		svc := services.NewEventService(services.EventServiceParams{
			Repo:    new(MockEventRepo),
			RBACSvc: new(MockRBACService),
			Logger:  nil,
		})
		assert.NotNil(t, svc)
	})
}
