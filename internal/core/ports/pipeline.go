// Package ports defines service and repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
)

// CreatePipelineOptions defines user-provided values for pipeline creation.
type CreatePipelineOptions struct {
	Name          string
	RepositoryURL string
	Branch        string
	WebhookSecret string
	Config        domain.PipelineConfig
}

// UpdatePipelineOptions defines mutable pipeline fields.
type UpdatePipelineOptions struct {
	Name          *string
	Branch        *string
	WebhookSecret *string
	Config        *domain.PipelineConfig
	Status        *domain.PipelineStatus
}

// TriggerBuildOptions defines metadata associated with a trigger action.
type TriggerBuildOptions struct {
	CommitHash  string
	TriggerType domain.BuildTriggerType
}

// WebhookTriggerOptions defines inbound webhook trigger inputs.
type WebhookTriggerOptions struct {
	PipelineID uuid.UUID
	Provider   string
	Event      string
	Signature  string
	DeliveryID string
	Payload    []byte
}

// PipelineRepository handles persistence of CI/CD pipelines, builds, and logs.
type PipelineRepository interface {
	CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) error
	GetPipelineByID(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error)
	GetPipeline(ctx context.Context, id, userID uuid.UUID) (*domain.Pipeline, error)
	ListPipelines(ctx context.Context, userID uuid.UUID) ([]*domain.Pipeline, error)
	UpdatePipeline(ctx context.Context, pipeline *domain.Pipeline) error
	DeletePipeline(ctx context.Context, id, userID uuid.UUID) error

	CreateBuild(ctx context.Context, build *domain.Build) error
	GetBuild(ctx context.Context, id, userID uuid.UUID) (*domain.Build, error)
	ListBuildsByPipeline(ctx context.Context, pipelineID, userID uuid.UUID) ([]*domain.Build, error)
	UpdateBuild(ctx context.Context, build *domain.Build) error

	CreateBuildStep(ctx context.Context, step *domain.BuildStep) error
	ListBuildSteps(ctx context.Context, buildID, userID uuid.UUID) ([]*domain.BuildStep, error)
	UpdateBuildStep(ctx context.Context, step *domain.BuildStep) error

	AppendBuildLog(ctx context.Context, log *domain.BuildLog) error
	ListBuildLogs(ctx context.Context, buildID, userID uuid.UUID, limit int) ([]*domain.BuildLog, error)
	ReserveWebhookDelivery(ctx context.Context, pipelineID uuid.UUID, provider, event, deliveryID string) (bool, error)
}

// PipelineService provides business logic for CI/CD pipeline lifecycle and execution.
type PipelineService interface {
	CreatePipeline(ctx context.Context, opts CreatePipelineOptions) (*domain.Pipeline, error)
	GetPipeline(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error)
	ListPipelines(ctx context.Context) ([]*domain.Pipeline, error)
	UpdatePipeline(ctx context.Context, id uuid.UUID, opts UpdatePipelineOptions) (*domain.Pipeline, error)
	DeletePipeline(ctx context.Context, id uuid.UUID) error

	TriggerBuild(ctx context.Context, pipelineID uuid.UUID, opts TriggerBuildOptions) (*domain.Build, error)
	TriggerBuildWebhook(ctx context.Context, opts WebhookTriggerOptions) (*domain.Build, error)
	GetBuild(ctx context.Context, buildID uuid.UUID) (*domain.Build, error)
	ListBuildsByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*domain.Build, error)
	ListBuildSteps(ctx context.Context, buildID uuid.UUID) ([]*domain.BuildStep, error)
	ListBuildLogs(ctx context.Context, buildID uuid.UUID, limit int) ([]*domain.BuildLog, error)
}
