package httphandlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const pipelineWebhookRoute = "/pipelines/:id/webhook/:provider"

type mockPipelineService struct{ mock.Mock }

func (m *mockPipelineService) CreatePipeline(ctx context.Context, opts ports.CreatePipelineOptions) (*domain.Pipeline, error) {
	args := m.Called(ctx, opts)
	r0, _ := args.Get(0).(*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *mockPipelineService) GetPipeline(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error) {
	args := m.Called(ctx, id)
	r0, _ := args.Get(0).(*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *mockPipelineService) ListPipelines(ctx context.Context) ([]*domain.Pipeline, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).([]*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *mockPipelineService) UpdatePipeline(ctx context.Context, id uuid.UUID, opts ports.UpdatePipelineOptions) (*domain.Pipeline, error) {
	args := m.Called(ctx, id, opts)
	r0, _ := args.Get(0).(*domain.Pipeline)
	return r0, args.Error(1)
}

func (m *mockPipelineService) DeletePipeline(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockPipelineService) TriggerBuild(ctx context.Context, pipelineID uuid.UUID, opts ports.TriggerBuildOptions) (*domain.Build, error) {
	args := m.Called(ctx, pipelineID, opts)
	r0, _ := args.Get(0).(*domain.Build)
	return r0, args.Error(1)
}

func (m *mockPipelineService) TriggerBuildWebhook(ctx context.Context, opts ports.WebhookTriggerOptions) (*domain.Build, error) {
	args := m.Called(ctx, opts)
	r0, _ := args.Get(0).(*domain.Build)
	return r0, args.Error(1)
}

func (m *mockPipelineService) GetBuild(ctx context.Context, buildID uuid.UUID) (*domain.Build, error) {
	args := m.Called(ctx, buildID)
	r0, _ := args.Get(0).(*domain.Build)
	return r0, args.Error(1)
}

func (m *mockPipelineService) ListBuildsByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*domain.Build, error) {
	args := m.Called(ctx, pipelineID)
	r0, _ := args.Get(0).([]*domain.Build)
	return r0, args.Error(1)
}

func (m *mockPipelineService) ListBuildSteps(ctx context.Context, buildID uuid.UUID) ([]*domain.BuildStep, error) {
	args := m.Called(ctx, buildID)
	r0, _ := args.Get(0).([]*domain.BuildStep)
	return r0, args.Error(1)
}

func (m *mockPipelineService) ListBuildLogs(ctx context.Context, buildID uuid.UUID, limit int) ([]*domain.BuildLog, error) {
	args := m.Called(ctx, buildID, limit)
	r0, _ := args.Get(0).([]*domain.BuildLog)
	return r0, args.Error(1)
}

func TestPipelineWebhookTriggerGitHubHeaders(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockPipelineService)
	h := NewPipelineHandler(svc)
	r := gin.New()
	r.POST(pipelineWebhookRoute, h.WebhookTrigger)

	pipelineID := uuid.New()
	body := []byte(`{"ref":"refs/heads/main","after":"abc123"}`)

	svc.On("TriggerBuildWebhook", mock.Anything, mock.MatchedBy(func(opts ports.WebhookTriggerOptions) bool {
		return opts.PipelineID == pipelineID &&
			opts.Provider == "github" &&
			opts.Event == "push" &&
			opts.Signature == "sha256=abc" &&
			opts.DeliveryID == "delivery-123" &&
			bytes.Equal(opts.Payload, body)
	})).Return(&domain.Build{ID: uuid.New(), PipelineID: pipelineID, Status: domain.BuildStatusQueued}, nil).Once()

	req, err := http.NewRequest(http.MethodPost, "/pipelines/"+pipelineID.String()+"/webhook/github", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", "sha256=abc")
	req.Header.Set("X-GitHub-Delivery", "delivery-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	svc.AssertExpectations(t)
}

func TestPipelineWebhookTriggerGitLabHeadersIgnored(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockPipelineService)
	h := NewPipelineHandler(svc)
	r := gin.New()
	r.POST(pipelineWebhookRoute, h.WebhookTrigger)

	pipelineID := uuid.New()
	body := []byte(`{"ref":"refs/heads/dev","after":"abc123"}`)

	svc.On("TriggerBuildWebhook", mock.Anything, mock.MatchedBy(func(opts ports.WebhookTriggerOptions) bool {
		return opts.PipelineID == pipelineID &&
			opts.Provider == "gitlab" &&
			opts.Event == "Push Hook" &&
			opts.Signature == "token-1" &&
			opts.DeliveryID == "uuid-evt-1" &&
			bytes.Equal(opts.Payload, body)
	})).Return(nil, nil).Once()

	req, err := http.NewRequest(http.MethodPost, "/pipelines/"+pipelineID.String()+"/webhook/gitlab", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", "token-1")
	req.Header.Set("X-Gitlab-Event-UUID", "uuid-evt-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), "ignored")
	svc.AssertExpectations(t)
}

func TestPipelineWebhookTriggerInvalidPipelineID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	svc := new(mockPipelineService)
	h := NewPipelineHandler(svc)
	r := gin.New()
	r.POST(pipelineWebhookRoute, h.WebhookTrigger)

	req, err := http.NewRequest(http.MethodPost, "/pipelines/not-a-uuid/webhook/github", bytes.NewReader([]byte(`{}`)))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertNotCalled(t, "TriggerBuildWebhook", mock.Anything, mock.Anything)
}
