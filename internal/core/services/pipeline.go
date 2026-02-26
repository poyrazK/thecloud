// Package services implements core business workflows.
package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

const pipelineBuildQueueName = "pipeline_build_queue"

// PipelineService manages CI/CD pipeline lifecycle and build orchestration.
type PipelineService struct {
	repo      ports.PipelineRepository
	taskQueue ports.TaskQueue
	eventSvc  ports.EventService
	auditSvc  ports.AuditService
}

// NewPipelineService constructs a PipelineService with its dependencies.
func NewPipelineService(repo ports.PipelineRepository, taskQueue ports.TaskQueue, eventSvc ports.EventService, auditSvc ports.AuditService) ports.PipelineService {
	return &PipelineService{
		repo:      repo,
		taskQueue: taskQueue,
		eventSvc:  eventSvc,
		auditSvc:  auditSvc,
	}
}

func (s *PipelineService) CreatePipeline(ctx context.Context, opts ports.CreatePipelineOptions) (*domain.Pipeline, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if opts.Name == "" || opts.RepositoryURL == "" {
		return nil, fmt.Errorf("name and repository_url are required")
	}
	if opts.Branch == "" {
		opts.Branch = "main"
	}

	now := time.Now()
	pipeline := &domain.Pipeline{
		ID:            uuid.New(),
		UserID:        userID,
		Name:          opts.Name,
		RepositoryURL: opts.RepositoryURL,
		Branch:        opts.Branch,
		WebhookSecret: opts.WebhookSecret,
		Config:        opts.Config,
		Status:        domain.PipelineStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreatePipeline(ctx, pipeline); err != nil {
		return nil, err
	}

	_ = s.eventSvc.RecordEvent(ctx, "PIPELINE_CREATED", pipeline.ID.String(), "PIPELINE", map[string]interface{}{"name": pipeline.Name})
	_ = s.auditSvc.Log(ctx, userID, "pipeline.create", "pipeline", pipeline.ID.String(), map[string]interface{}{"name": pipeline.Name, "branch": pipeline.Branch})

	return pipeline, nil
}

func (s *PipelineService) GetPipeline(ctx context.Context, id uuid.UUID) (*domain.Pipeline, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	pipeline, err := s.repo.GetPipeline(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, fmt.Errorf("pipeline not found")
	}

	return pipeline, nil
}

func (s *PipelineService) ListPipelines(ctx context.Context) ([]*domain.Pipeline, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	return s.repo.ListPipelines(ctx, userID)
}

func (s *PipelineService) UpdatePipeline(ctx context.Context, id uuid.UUID, opts ports.UpdatePipelineOptions) (*domain.Pipeline, error) {
	pipeline, err := s.GetPipeline(ctx, id)
	if err != nil {
		return nil, err
	}

	if opts.Name != nil {
		pipeline.Name = *opts.Name
	}
	if opts.Branch != nil {
		pipeline.Branch = *opts.Branch
	}
	if opts.WebhookSecret != nil {
		pipeline.WebhookSecret = *opts.WebhookSecret
	}
	if opts.Config != nil {
		pipeline.Config = *opts.Config
	}
	if opts.Status != nil {
		pipeline.Status = *opts.Status
	}
	pipeline.UpdatedAt = time.Now()

	if err := s.repo.UpdatePipeline(ctx, pipeline); err != nil {
		return nil, err
	}

	_ = s.eventSvc.RecordEvent(ctx, "PIPELINE_UPDATED", pipeline.ID.String(), "PIPELINE", nil)
	_ = s.auditSvc.Log(ctx, pipeline.UserID, "pipeline.update", "pipeline", pipeline.ID.String(), map[string]interface{}{})

	return pipeline, nil
}

func (s *PipelineService) DeletePipeline(ctx context.Context, id uuid.UUID) error {
	pipeline, err := s.GetPipeline(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeletePipeline(ctx, id, pipeline.UserID); err != nil {
		return err
	}

	_ = s.eventSvc.RecordEvent(ctx, "PIPELINE_DELETED", pipeline.ID.String(), "PIPELINE", nil)
	_ = s.auditSvc.Log(ctx, pipeline.UserID, "pipeline.delete", "pipeline", pipeline.ID.String(), map[string]interface{}{"name": pipeline.Name})

	return nil
}

func (s *PipelineService) TriggerBuild(ctx context.Context, pipelineID uuid.UUID, opts ports.TriggerBuildOptions) (*domain.Build, error) {
	pipeline, err := s.GetPipeline(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	if pipeline.Status != domain.PipelineStatusActive {
		return nil, fmt.Errorf("pipeline is not active")
	}

	if opts.TriggerType == "" {
		opts.TriggerType = domain.BuildTriggerManual
	}

	return s.createAndQueueBuild(ctx, pipeline, opts.CommitHash, opts.TriggerType)
}

func (s *PipelineService) TriggerBuildWebhook(ctx context.Context, opts ports.WebhookTriggerOptions) (*domain.Build, error) {
	pipeline, err := s.repo.GetPipelineByID(ctx, opts.PipelineID)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		return nil, fmt.Errorf("pipeline not found")
	}
	if pipeline.Status != domain.PipelineStatusActive {
		return nil, fmt.Errorf("pipeline is not active")
	}
	if pipeline.WebhookSecret == "" {
		return nil, fmt.Errorf("pipeline webhook secret is not configured")
	}

	provider := strings.ToLower(strings.TrimSpace(opts.Provider))
	if err := validateWebhookSignature(provider, pipeline.WebhookSecret, opts.Signature, opts.Payload); err != nil {
		return nil, err
	}

	deliveryID := strings.TrimSpace(opts.DeliveryID)
	if deliveryID != "" {
		reserved, err := s.repo.ReserveWebhookDelivery(ctx, pipeline.ID, provider, strings.TrimSpace(opts.Event), deliveryID)
		if err != nil {
			return nil, err
		}
		if !reserved {
			return nil, nil
		}
	}

	branch, commitHash, ok, err := extractWebhookBuildInfo(provider, strings.TrimSpace(opts.Event), opts.Payload)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	if branch != "" && branch != pipeline.Branch {
		return nil, nil
	}

	webhookCtx := appcontext.WithUserID(ctx, pipeline.UserID)
	return s.createAndQueueBuild(webhookCtx, pipeline, commitHash, domain.BuildTriggerWebhook)
}

func (s *PipelineService) createAndQueueBuild(ctx context.Context, pipeline *domain.Pipeline, commitHash string, triggerType domain.BuildTriggerType) (*domain.Build, error) {

	now := time.Now()
	build := &domain.Build{
		ID:          uuid.New(),
		PipelineID:  pipeline.ID,
		UserID:      pipeline.UserID,
		CommitHash:  commitHash,
		TriggerType: triggerType,
		Status:      domain.BuildStatusQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateBuild(ctx, build); err != nil {
		return nil, err
	}

	job := domain.BuildJob{
		BuildID:    build.ID,
		PipelineID: build.PipelineID,
		UserID:     build.UserID,
		CommitHash: build.CommitHash,
		Trigger:    build.TriggerType,
	}
	if err := s.taskQueue.Enqueue(ctx, pipelineBuildQueueName, job); err != nil {
		return nil, err
	}

	_ = s.eventSvc.RecordEvent(ctx, "PIPELINE_BUILD_QUEUED", build.ID.String(), "PIPELINE_BUILD", map[string]interface{}{"pipeline_id": pipeline.ID})
	_ = s.auditSvc.Log(ctx, build.UserID, "pipeline.run", "pipeline_build", build.ID.String(), map[string]interface{}{"pipeline_id": pipeline.ID})

	return build, nil
}

func validateWebhookSignature(provider, secret, signature string, payload []byte) error {
	sig := strings.TrimSpace(signature)
	if sig == "" {
		return fmt.Errorf("missing webhook signature")
	}

	switch provider {
	case "github":
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write(payload)
		expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if subtle.ConstantTimeCompare([]byte(expected), []byte(sig)) != 1 {
			return fmt.Errorf("invalid github webhook signature")
		}
		return nil
	case "gitlab":
		if subtle.ConstantTimeCompare([]byte(secret), []byte(sig)) != 1 {
			return fmt.Errorf("invalid gitlab webhook token")
		}
		return nil
	default:
		return fmt.Errorf("unsupported webhook provider")
	}
}

func extractWebhookBuildInfo(provider, event string, payload []byte) (string, string, bool, error) {
	switch provider {
	case "github":
		if event != "push" {
			return "", "", false, nil
		}
		var body struct {
			Ref   string `json:"ref"`
			After string `json:"after"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return "", "", false, fmt.Errorf("invalid github payload: %w", err)
		}
		return normalizeBranch(body.Ref), body.After, true, nil
	case "gitlab":
		if event != "Push Hook" {
			return "", "", false, nil
		}
		var body struct {
			Ref         string `json:"ref"`
			After       string `json:"after"`
			CheckoutSHA string `json:"checkout_sha"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return "", "", false, fmt.Errorf("invalid gitlab payload: %w", err)
		}
		commit := body.CheckoutSHA
		if commit == "" {
			commit = body.After
		}
		return normalizeBranch(body.Ref), commit, true, nil
	default:
		return "", "", false, fmt.Errorf("unsupported webhook provider")
	}
}

func normalizeBranch(ref string) string {
	trimmed := strings.TrimSpace(ref)
	trimmed = strings.TrimPrefix(trimmed, "refs/heads/")
	return trimmed
}

func (s *PipelineService) GetBuild(ctx context.Context, buildID uuid.UUID) (*domain.Build, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	build, err := s.repo.GetBuild(ctx, buildID, userID)
	if err != nil {
		return nil, err
	}
	if build == nil {
		return nil, fmt.Errorf("build not found")
	}

	return build, nil
}

func (s *PipelineService) ListBuildsByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*domain.Build, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	return s.repo.ListBuildsByPipeline(ctx, pipelineID, userID)
}

func (s *PipelineService) ListBuildSteps(ctx context.Context, buildID uuid.UUID) ([]*domain.BuildStep, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	return s.repo.ListBuildSteps(ctx, buildID, userID)
}

func (s *PipelineService) ListBuildLogs(ctx context.Context, buildID uuid.UUID, limit int) ([]*domain.BuildLog, error) {
	userID := appcontext.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		return nil, fmt.Errorf("unauthorized")
	}

	return s.repo.ListBuildLogs(ctx, buildID, userID, limit)
}
