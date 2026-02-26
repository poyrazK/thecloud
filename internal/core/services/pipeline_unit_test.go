package services_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testWebhookSecret = "top-secret"

type MockPipelineRepository struct{ mock.Mock }

func (m *MockPipelineRepository) CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	return m.Called(ctx, pipeline).Error(0)
}

func (m *MockPipelineRepository) GetPipelineByID(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) GetPipeline(ctx context.Context, id, userID uuid.UUID) (*domain.Pipeline, error) {
	args := m.Called(ctx, id, userID)
	r0, _ := args.Get(0).(*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) ListPipelines(ctx context.Context, userID uuid.UUID) ([]*domain.Pipeline, error) {
	args := m.Called(ctx, userID)
	r0, _ := args.Get(0).([]*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) UpdatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	return m.Called(ctx, pipeline).Error(0)
}

func (m *MockPipelineRepository) DeletePipeline(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *MockPipelineRepository) CreateBuild(ctx context.Context, build *domain.Build) error {
	return m.Called(ctx, build).Error(0)
}

func (m *MockPipelineRepository) GetBuild(ctx context.Context, id, userID uuid.UUID) (*domain.Build, error) {
	args := m.Called(ctx, id, userID)
	r0, _ := args.Get(0).(*domain.Build)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) ListBuildsByPipeline(ctx context.Context, pipelineID, userID uuid.UUID) ([]*domain.Build, error) {
	args := m.Called(ctx, pipelineID, userID)
	r0, _ := args.Get(0).([]*domain.Build)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) UpdateBuild(ctx context.Context, build *domain.Build) error {
	return m.Called(ctx, build).Error(0)
}

func (m *MockPipelineRepository) CreateBuildStep(ctx context.Context, step *domain.BuildStep) error {
	return m.Called(ctx, step).Error(0)
}

func (m *MockPipelineRepository) ListBuildSteps(ctx context.Context, buildID, userID uuid.UUID) ([]*domain.BuildStep, error) {
	args := m.Called(ctx, buildID, userID)
	r0, _ := args.Get(0).([]*domain.BuildStep)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) UpdateBuildStep(ctx context.Context, step *domain.BuildStep) error {
	return m.Called(ctx, step).Error(0)
}

func (m *MockPipelineRepository) AppendBuildLog(ctx context.Context, log *domain.BuildLog) error {
	return m.Called(ctx, log).Error(0)
}

func (m *MockPipelineRepository) ListBuildLogs(ctx context.Context, buildID, userID uuid.UUID, limit int) ([]*domain.BuildLog, error) {
	args := m.Called(ctx, buildID, userID, limit)
	r0, _ := args.Get(0).([]*domain.BuildLog)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) ReserveWebhookDelivery(ctx context.Context, pipelineID uuid.UUID, provider, event, deliveryID string) (bool, error) {
	args := m.Called(ctx, pipelineID, provider, event, deliveryID)
	return args.Bool(0), args.Error(1)
}

func githubSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestPipelineServiceTriggerBuildWebhookInvalidGitHubSignature(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc)
	pipelineID := uuid.New()

	repo.On("GetPipelineByID", mock.Anything, pipelineID).Return(&domain.Pipeline{
		ID:            pipelineID,
		UserID:        uuid.New(),
		Branch:        "main",
		WebhookSecret: testWebhookSecret,
		Status:        domain.PipelineStatusActive,
	}, nil).Once()

	_, err := svc.TriggerBuildWebhook(context.Background(), ports.WebhookTriggerOptions{
		PipelineID: pipelineID,
		Provider:   "github",
		Event:      "push",
		Signature:  "sha256=bad",
		Payload:    []byte(`{"ref":"refs/heads/main","after":"abc123"}`),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid github webhook signature")
	repo.AssertExpectations(t)
}

func TestPipelineServiceTriggerBuildWebhookDuplicateDeliveryIgnored(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc)
	pipelineID := uuid.New()
	payload := []byte(`{"ref":"refs/heads/main","after":"abc123"}`)

	repo.On("GetPipelineByID", mock.Anything, pipelineID).Return(&domain.Pipeline{
		ID:            pipelineID,
		UserID:        uuid.New(),
		Branch:        "main",
		WebhookSecret: testWebhookSecret,
		Status:        domain.PipelineStatusActive,
	}, nil).Once()
	repo.On("ReserveWebhookDelivery", mock.Anything, pipelineID, "github", "push", "delivery-1").Return(false, nil).Once()

	build, err := svc.TriggerBuildWebhook(context.Background(), ports.WebhookTriggerOptions{
		PipelineID:  pipelineID,
		Provider:    "github",
		Event:       "push",
		Signature:   githubSignature(testWebhookSecret, payload),
		DeliveryID:  "delivery-1",
		Payload:     payload,
	})

	require.NoError(t, err)
	assert.Nil(t, build)
	repo.AssertExpectations(t)
}

func TestPipelineServiceTriggerBuildWebhookBranchMismatchIgnored(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc)
	pipelineID := uuid.New()
	payload := []byte(`{"ref":"refs/heads/dev","after":"abc123"}`)

	repo.On("GetPipelineByID", mock.Anything, pipelineID).Return(&domain.Pipeline{
		ID:            pipelineID,
		UserID:        uuid.New(),
		Branch:        "main",
		WebhookSecret: testWebhookSecret,
		Status:        domain.PipelineStatusActive,
	}, nil).Once()
	repo.On("ReserveWebhookDelivery", mock.Anything, pipelineID, "github", "push", "delivery-2").Return(true, nil).Once()

	build, err := svc.TriggerBuildWebhook(context.Background(), ports.WebhookTriggerOptions{
		PipelineID:  pipelineID,
		Provider:    "github",
		Event:       "push",
		Signature:   githubSignature(testWebhookSecret, payload),
		DeliveryID:  "delivery-2",
		Payload:     payload,
	})

	require.NoError(t, err)
	assert.Nil(t, build)
	repo.AssertExpectations(t)
}

func TestPipelineServiceTriggerBuildWebhookSuccessQueuesBuild(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc)
	userID := uuid.New()
	pipelineID := uuid.New()
	payload := []byte(`{"ref":"refs/heads/main","after":"abc123"}`)

	repo.On("GetPipelineByID", mock.Anything, pipelineID).Return(&domain.Pipeline{
		ID:            pipelineID,
		UserID:        userID,
		Branch:        "main",
		WebhookSecret: testWebhookSecret,
		Status:        domain.PipelineStatusActive,
	}, nil).Once()
	repo.On("ReserveWebhookDelivery", mock.Anything, pipelineID, "github", "push", "delivery-3").Return(true, nil).Once()
	repo.On("CreateBuild", mock.Anything, mock.MatchedBy(func(b *domain.Build) bool {
		return b != nil && b.PipelineID == pipelineID && b.UserID == userID && b.TriggerType == domain.BuildTriggerWebhook && b.CommitHash == "abc123"
	})).Return(nil).Once()
	taskQueue.On("Enqueue", mock.Anything, "pipeline_build_queue", mock.Anything).Return(nil).Once()
	eventSvc.On("RecordEvent", mock.Anything, "PIPELINE_BUILD_QUEUED", mock.Anything, "PIPELINE_BUILD", mock.Anything).Return(nil).Once()
	auditSvc.On("Log", mock.Anything, userID, "pipeline.run", "pipeline_build", mock.Anything, mock.Anything).Return(nil).Once()

	build, err := svc.TriggerBuildWebhook(context.Background(), ports.WebhookTriggerOptions{
		PipelineID:  pipelineID,
		Provider:    "github",
		Event:       "push",
		Signature:   githubSignature(testWebhookSecret, payload),
		DeliveryID:  "delivery-3",
		Payload:     payload,
	})

	require.NoError(t, err)
	require.NotNil(t, build)
	assert.Equal(t, domain.BuildTriggerWebhook, build.TriggerType)
	assert.Equal(t, "abc123", build.CommitHash)
	assert.Equal(t, domain.BuildStatusQueued, build.Status)

	repo.AssertExpectations(t)
	taskQueue.AssertExpectations(t)
	eventSvc.AssertExpectations(t)
	auditSvc.AssertExpectations(t)
}
