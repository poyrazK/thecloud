package services_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEventRepository struct {
	mock.Mock
}

func (m *mockEventRepository) Create(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventRepository) List(ctx context.Context, limit int) ([]*domain.Event, error) {
	args := m.Called(ctx, limit)
	r0, _ := args.Get(0).([]*domain.Event)
	return r0, args.Error(1)
}

func TestEventService_RecordEvent(t *testing.T) {
	repo := new(mockEventRepository)
	logger := slog.Default()
	svc := services.NewEventService(repo, nil, logger)

	ctx := appcontext.WithUserID(context.Background(), uuid.New())
	action := "test.action"
	resourceID := uuid.New().String()
	resourceType := "instance"
	metadata := map[string]interface{}{"foo": "bar"}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(e *domain.Event) bool {
		return e.Action == action && e.ResourceID == resourceID && e.ResourceType == resourceType
	})).Return(nil)

	err := svc.RecordEvent(ctx, action, resourceID, resourceType, metadata)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestEventService_ListEvents(t *testing.T) {
	repo := new(mockEventRepository)
	svc := services.NewEventService(repo, nil, slog.Default())

	ctx := context.Background()
	limit := 10
	expectedEvents := []*domain.Event{{ID: uuid.New()}}

	repo.On("List", ctx, limit).Return(expectedEvents, nil)

	events, err := svc.ListEvents(ctx, limit)
	assert.NoError(t, err)
	assert.Equal(t, expectedEvents, events)
	repo.AssertExpectations(t)
}

func TestEventService_ListEventsDefaultLimit(t *testing.T) {
	repo := new(mockEventRepository)
	svc := services.NewEventService(repo, nil, slog.Default())

	ctx := context.Background()
	repo.On("List", ctx, 50).Return([]*domain.Event{}, nil)

	_, err := svc.ListEvents(ctx, 0)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
