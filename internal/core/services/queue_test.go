package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockQueueRepository struct {
	mock.Mock
}

func (m *mockQueueRepository) Create(ctx context.Context, q *domain.Queue) error {
	args := m.Called(ctx, q)
	return args.Error(0)
}

func (m *mockQueueRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Queue, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}

func (m *mockQueueRepository) GetByName(ctx context.Context, name string, userID uuid.UUID) (*domain.Queue, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Queue), args.Error(1)
}

func (m *mockQueueRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Queue, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Queue), args.Error(1)
}

func (m *mockQueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockQueueRepository) PurgeMessages(ctx context.Context, queueID uuid.UUID) (int64, error) {
	args := m.Called(ctx, queueID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockQueueRepository) SendMessage(ctx context.Context, queueID uuid.UUID, body string) (*domain.Message, error) {
	args := m.Called(ctx, queueID, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Message), args.Error(1)
}

func (m *mockQueueRepository) ReceiveMessages(ctx context.Context, queueID uuid.UUID, maxMessages int, visibilityTimeout int) ([]*domain.Message, error) {
	args := m.Called(ctx, queueID, maxMessages, visibilityTimeout)
	return args.Get(0).([]*domain.Message), args.Error(1)
}

func (m *mockQueueRepository) DeleteMessage(ctx context.Context, queueID uuid.UUID, receiptHandle string) error {
	args := m.Called(ctx, queueID, receiptHandle)
	return args.Error(0)
}

func TestQueueServiceCreateQueue(t *testing.T) {
	repo := new(mockQueueRepository)
	eventSvc := new(mockEventService)
	auditSvc := new(mockAuditService)
	svc := NewQueueService(repo, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	t.Run("success", func(t *testing.T) {
		repo.On("GetByName", mock.Anything, "test-queue", userID).Return(nil, nil).Once()
		repo.On("Create", mock.Anything, mock.MatchedBy(func(q *domain.Queue) bool {
			return q.Name == "test-queue" && q.UserID == userID
		})).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "QUEUE_CREATED", mock.Anything, "QUEUE", mock.Anything).Return(nil).Once()
		auditSvc.On("Log", mock.Anything, userID, "queue.create", "queue", mock.Anything, mock.Anything).Return(nil).Once()

		opts := &ports.CreateQueueOptions{}
		q, err := svc.CreateQueue(ctx, "test-queue", opts)

		assert.NoError(t, err)
		assert.NotNil(t, q)
		assert.Equal(t, "test-queue", q.Name)
		repo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		_, err := svc.CreateQueue(context.Background(), "test-queue", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("already exists", func(t *testing.T) {
		repo.On("GetByName", mock.Anything, "existing", userID).Return(&domain.Queue{ID: uuid.New()}, nil).Once()

		_, err := svc.CreateQueue(ctx, "existing", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestQueueServiceSendMessage(t *testing.T) {
	repo := new(mockQueueRepository)
	eventSvc := new(mockEventService)
	auditSvc := new(mockAuditService)
	svc := NewQueueService(repo, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	qID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, qID, userID).Return(&domain.Queue{
			ID:             qID,
			UserID:         userID,
			MaxMessageSize: 256 * 1024,
		}, nil).Once()
		repo.On("SendMessage", mock.Anything, qID, "hello").Return(&domain.Message{ID: uuid.New(), Body: "hello"}, nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "MESSAGE_SENT", mock.Anything, "MESSAGE", mock.Anything).Return(nil).Once()

		msg, err := svc.SendMessage(ctx, qID, "hello")
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		repo.AssertExpectations(t)
	})
}

func TestQueueServiceReceiveMessages(t *testing.T) {
	repo := new(mockQueueRepository)
	eventSvc := new(mockEventService)
	auditSvc := new(mockAuditService)
	svc := NewQueueService(repo, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	qID := uuid.New()

	t.Run("success", func(t *testing.T) {
		repo.On("GetByID", mock.Anything, qID, userID).Return(&domain.Queue{
			ID:                qID,
			UserID:            userID,
			VisibilityTimeout: 30,
		}, nil).Once()
		msgs := []*domain.Message{{ID: uuid.New(), Body: "msg1"}}
		repo.On("ReceiveMessages", mock.Anything, qID, 10, 30).Return(msgs, nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, "MESSAGE_RECEIVED", mock.Anything, "MESSAGE", mock.Anything).Return(nil).Once()

		result, err := svc.ReceiveMessages(ctx, qID, 10)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

func TestQueueServiceDeleteQueue(t *testing.T) {
	repo := new(mockQueueRepository)
	eventSvc := new(mockEventService)
	auditSvc := new(mockAuditService)
	svc := NewQueueService(repo, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	qID := uuid.New()

	repo.On("GetByID", mock.Anything, qID, userID).Return(&domain.Queue{ID: qID, UserID: userID, Name: "to-delete"}, nil).Once()
	repo.On("Delete", mock.Anything, qID).Return(nil).Once()
	eventSvc.On("RecordEvent", mock.Anything, "QUEUE_DELETED", qID.String(), "QUEUE", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, userID, "queue.delete", "queue", qID.String(), mock.Anything).Return(nil).Once()

	err := svc.DeleteQueue(ctx, qID)
	assert.NoError(t, err)
}

func TestQueueServiceListQueues(t *testing.T) {
	repo := new(mockQueueRepository)
	eventSvc := new(mockEventService)
	auditSvc := new(mockAuditService)
	svc := NewQueueService(repo, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)

	repo.On("List", mock.Anything, userID).Return([]*domain.Queue{{ID: uuid.New()}}, nil).Once()
	queues, err := svc.ListQueues(ctx)
	assert.NoError(t, err)
	assert.Len(t, queues, 1)
}

func TestQueueServiceDeleteMessage(t *testing.T) {
	repo := new(mockQueueRepository)
	eventSvc := new(mockEventService)
	auditSvc := new(mockAuditService)
	svc := NewQueueService(repo, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	qID := uuid.New()
	receipt := "receipt-1"

	queue := &domain.Queue{ID: qID, UserID: userID}
	repo.On("GetByID", mock.Anything, qID, userID).Return(queue, nil).Once()
	repo.On("DeleteMessage", mock.Anything, qID, receipt).Return(nil).Once()
	eventSvc.On("RecordEvent", mock.Anything, "MESSAGE_DELETED", receipt, "MESSAGE", mock.Anything).Return(nil).Once()

	err := svc.DeleteMessage(ctx, qID, receipt)
	assert.NoError(t, err)
}

func TestQueueServicePurgeQueue(t *testing.T) {
	repo := new(mockQueueRepository)
	eventSvc := new(mockEventService)
	auditSvc := new(mockAuditService)
	svc := NewQueueService(repo, eventSvc, auditSvc)

	userID := uuid.New()
	ctx := appcontext.WithUserID(context.Background(), userID)
	qID := uuid.New()

	queue := &domain.Queue{ID: qID, UserID: userID}
	repo.On("GetByID", mock.Anything, qID, userID).Return(queue, nil).Once()
	repo.On("PurgeMessages", mock.Anything, qID).Return(int64(2), nil).Once()
	eventSvc.On("RecordEvent", mock.Anything, "QUEUE_PURGED", qID.String(), "QUEUE", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, userID, "queue.purge", "queue", qID.String(), mock.Anything).Return(nil).Once()

	err := svc.PurgeQueue(ctx, qID)
	assert.NoError(t, err)
}
