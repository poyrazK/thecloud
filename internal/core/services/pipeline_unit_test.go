package services_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
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

func (m *MockPipelineRepository) GetPipeline(ctx context.Context, id, tenantID uuid.UUID) (*domain.Pipeline, error) {
	args := m.Called(ctx, id, tenantID)
	r0, _ := args.Get(0).(*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) ListPipelines(ctx context.Context, tenantID uuid.UUID) ([]*domain.Pipeline, error) {
	args := m.Called(ctx, tenantID)
	r0, _ := args.Get(0).([]*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) UpdatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	return m.Called(ctx, pipeline).Error(0)
}

func (m *MockPipelineRepository) DeletePipeline(ctx context.Context, id, tenantID uuid.UUID) error {
	return m.Called(ctx, id, tenantID).Error(0)
}

func (m *MockPipelineRepository) CreateBuild(ctx context.Context, build *domain.Build) error {
	return m.Called(ctx, build).Error(0)
}

func (m *MockPipelineRepository) GetBuild(ctx context.Context, id, tenantID uuid.UUID) (*domain.Build, error) {
	args := m.Called(ctx, id, tenantID)
	r0, _ := args.Get(0).(*domain.Build)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) ListBuildsByPipeline(ctx context.Context, pipelineID, tenantID uuid.UUID) ([]*domain.Build, error) {
	args := m.Called(ctx, pipelineID, tenantID)
	r0, _ := args.Get(0).([]*domain.Build)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) UpdateBuild(ctx context.Context, build *domain.Build) error {
	return m.Called(ctx, build).Error(0)
}

func (m *MockPipelineRepository) CreateBuildStep(ctx context.Context, step *domain.BuildStep) error {
	return m.Called(ctx, step).Error(0)
}

func (m *MockPipelineRepository) ListBuildSteps(ctx context.Context, buildID, tenantID uuid.UUID) ([]*domain.BuildStep, error) {
	args := m.Called(ctx, buildID, tenantID)
	r0, _ := args.Get(0).([]*domain.BuildStep)
	return r0, args.Error(1)
}

func (m *MockPipelineRepository) UpdateBuildStep(ctx context.Context, step *domain.BuildStep) error {
	return m.Called(ctx, step).Error(0)
}

func (m *MockPipelineRepository) AppendBuildLog(ctx context.Context, log *domain.BuildLog) error {
	return m.Called(ctx, log).Error(0)
}

func (m *MockPipelineRepository) ListBuildLogs(ctx context.Context, buildID, tenantID uuid.UUID, limit int) ([]*domain.BuildLog, error) {
	args := m.Called(ctx, buildID, tenantID, limit)
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

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())
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

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())
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
		PipelineID: pipelineID,
		Provider:   "github",
		Event:      "push",
		Signature:  githubSignature(testWebhookSecret, payload),
		DeliveryID: "delivery-1",
		Payload:    payload,
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

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())
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
		PipelineID: pipelineID,
		Provider:   "github",
		Event:      "push",
		Signature:  githubSignature(testWebhookSecret, payload),
		DeliveryID: "delivery-2",
		Payload:    payload,
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

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())
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
		PipelineID: pipelineID,
		Provider:   "github",
		Event:      "push",
		Signature:  githubSignature(testWebhookSecret, payload),
		DeliveryID: "delivery-3",
		Payload:    payload,
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

func TestPipelineService_CreatePipeline(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		repo.On("CreatePipeline", mock.Anything, mock.Anything).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
		auditSvc.On("Log", mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		pipeline, err := svc.CreatePipeline(ctx, ports.CreatePipelineOptions{
			Name:          "my-pipeline",
			RepositoryURL: "https://github.com/user/repo",
			Branch:        "main",
		})

		require.NoError(t, err)
		assert.NotNil(t, pipeline)
		assert.Equal(t, "my-pipeline", pipeline.Name)
		repo.AssertExpectations(t)
	})

	t.Run("MissingName", func(t *testing.T) {
		_, err := svc.CreatePipeline(ctx, ports.CreatePipelineOptions{
			RepositoryURL: "https://github.com/user/repo",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("MissingRepositoryURL", func(t *testing.T) {
		_, err := svc.CreatePipeline(ctx, ports.CreatePipelineOptions{
			Name: "my-pipeline",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestPipelineService_GetPipeline(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		pipelineID := uuid.New()
		pipeline := &domain.Pipeline{ID: pipelineID, UserID: userID, TenantID: tenantID, Name: "test"}

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(pipeline, nil).Once()

		res, err := svc.GetPipeline(ctx, pipelineID)
		require.NoError(t, err)
		assert.Equal(t, pipelineID, res.ID)
		repo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		pipelineID := uuid.New()

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(nil, nil).Once()

		_, err := svc.GetPipeline(ctx, pipelineID)
		require.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_ListPipelines(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		pipelines := []*domain.Pipeline{
			{ID: uuid.New(), Name: "pipeline-1"},
			{ID: uuid.New(), Name: "pipeline-2"},
		}

		repo.On("ListPipelines", mock.Anything, mock.Anything).Return(pipelines, nil).Once()

		res, err := svc.ListPipelines(ctx)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_UpdatePipeline(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		pipelineID := uuid.New()
		pipeline := &domain.Pipeline{ID: pipelineID, UserID: userID, TenantID: tenantID, Name: "old-name", Status: domain.PipelineStatusActive}

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(pipeline, nil).Once()
		repo.On("UpdatePipeline", mock.Anything, mock.Anything).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
		auditSvc.On("Log", mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		newName := "new-name"
		res, err := svc.UpdatePipeline(ctx, pipelineID, ports.UpdatePipelineOptions{Name: &newName})
		require.NoError(t, err)
		assert.Equal(t, "new-name", res.Name)
		repo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		pipelineID := uuid.New()

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(nil, nil).Once()

		_, err := svc.UpdatePipeline(ctx, pipelineID, ports.UpdatePipelineOptions{Name: stringPtr("new-name")})
		require.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_DeletePipeline(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		pipelineID := uuid.New()
		pipeline := &domain.Pipeline{ID: pipelineID, UserID: userID, TenantID: tenantID, Name: "test"}

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(pipeline, nil).Once()
		repo.On("DeletePipeline", mock.Anything, pipelineID, mock.Anything).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
		auditSvc.On("Log", mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		err := svc.DeletePipeline(ctx, pipelineID)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		pipelineID := uuid.New()

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(nil, nil).Once()

		err := svc.DeletePipeline(ctx, pipelineID)
		require.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_TriggerBuild(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		pipelineID := uuid.New()
		pipeline := &domain.Pipeline{ID: pipelineID, UserID: userID, TenantID: tenantID, Status: domain.PipelineStatusActive}

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(pipeline, nil).Once()
		repo.On("CreateBuild", mock.Anything, mock.MatchedBy(func(b *domain.Build) bool {
			return b != nil && b.PipelineID == pipelineID && b.CommitHash == "abc123"
		})).Return(nil).Once()
		taskQueue.On("Enqueue", mock.Anything, "pipeline_build_queue", mock.MatchedBy(func(job domain.BuildJob) bool {
			return job.PipelineID == pipelineID && job.CommitHash == "abc123"
		})).Return(nil).Once()
		eventSvc.On("RecordEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
		auditSvc.On("Log", mock.Anything, userID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		build, err := svc.TriggerBuild(ctx, pipelineID, ports.TriggerBuildOptions{
			CommitHash: "abc123",
		})

		require.NoError(t, err)
		assert.NotNil(t, build)
		assert.Equal(t, "abc123", build.CommitHash)
		repo.AssertExpectations(t)
		taskQueue.AssertExpectations(t)
	})

	t.Run("PipelineNotActive", func(t *testing.T) {
		pipelineID := uuid.New()
		pipeline := &domain.Pipeline{ID: pipelineID, UserID: userID, TenantID: tenantID, Status: domain.PipelineStatusPaused}

		repo.On("GetPipeline", mock.Anything, pipelineID, mock.Anything).Return(pipeline, nil).Once()

		_, err := svc.TriggerBuild(ctx, pipelineID, ports.TriggerBuildOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_GetBuild(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		buildID := uuid.New()
		build := &domain.Build{ID: buildID, UserID: userID, TenantID: tenantID}

		repo.On("GetBuild", mock.Anything, buildID, mock.Anything).Return(build, nil).Once()

		res, err := svc.GetBuild(ctx, buildID)
		require.NoError(t, err)
		assert.Equal(t, buildID, res.ID)
		repo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		buildID := uuid.New()

		repo.On("GetBuild", mock.Anything, buildID, mock.Anything).Return(nil, nil).Once()

		_, err := svc.GetBuild(ctx, buildID)
		require.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_ListBuildsByPipeline(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		pipelineID := uuid.New()
		builds := []*domain.Build{
			{ID: uuid.New()},
			{ID: uuid.New()},
		}

		repo.On("ListBuildsByPipeline", mock.Anything, pipelineID, mock.Anything).Return(builds, nil).Once()

		res, err := svc.ListBuildsByPipeline(ctx, pipelineID)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_ListBuildSteps(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		buildID := uuid.New()
		steps := []*domain.BuildStep{
			{ID: uuid.New()},
			{ID: uuid.New()},
		}

		repo.On("ListBuildSteps", mock.Anything, buildID, mock.Anything).Return(steps, nil).Once()

		res, err := svc.ListBuildSteps(ctx, buildID)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_ListBuildLogs(t *testing.T) {
	repo := new(MockPipelineRepository)
	taskQueue := new(MockTaskQueue)
	eventSvc := new(MockEventService)
	auditSvc := new(MockAuditService)

	svc := services.NewPipelineService(repo, taskQueue, eventSvc, auditSvc, slog.Default())

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	ctx = appcontext.WithUserID(ctx, userID)
	ctx = appcontext.WithTenantID(ctx, tenantID)

	t.Run("Success", func(t *testing.T) {
		buildID := uuid.New()
		logs := []*domain.BuildLog{
			{ID: uuid.New()},
			{ID: uuid.New()},
		}

		repo.On("ListBuildLogs", mock.Anything, buildID, mock.Anything, 100).Return(logs, nil).Once()

		res, err := svc.ListBuildLogs(ctx, buildID, 100)
		require.NoError(t, err)
		assert.Len(t, res, 2)
		repo.AssertExpectations(t)
	})
}

func TestPipelineService_Unit(t *testing.T) {
	t.Run("TriggerBuildWebhook_InvalidSignature", TestPipelineServiceTriggerBuildWebhookInvalidGitHubSignature)
	t.Run("TriggerBuildWebhook_DuplicateDelivery", TestPipelineServiceTriggerBuildWebhookDuplicateDeliveryIgnored)
	t.Run("TriggerBuildWebhook_BranchMismatch", TestPipelineServiceTriggerBuildWebhookBranchMismatchIgnored)
	t.Run("TriggerBuildWebhook_Success", TestPipelineServiceTriggerBuildWebhookSuccessQueuesBuild)
	t.Run("CreatePipeline", TestPipelineService_CreatePipeline)
	t.Run("GetPipeline", TestPipelineService_GetPipeline)
	t.Run("ListPipelines", TestPipelineService_ListPipelines)
	t.Run("UpdatePipeline", TestPipelineService_UpdatePipeline)
	t.Run("DeletePipeline", TestPipelineService_DeletePipeline)
	t.Run("TriggerBuild", TestPipelineService_TriggerBuild)
	t.Run("GetBuild", TestPipelineService_GetBuild)
	t.Run("ListBuildsByPipeline", TestPipelineService_ListBuildsByPipeline)
	t.Run("ListBuildSteps", TestPipelineService_ListBuildSteps)
	t.Run("ListBuildLogs", TestPipelineService_ListBuildLogs)
}

func stringPtr(s string) *string {
	return &s
}
