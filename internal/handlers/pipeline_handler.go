// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// PipelineHandler handles CI/CD pipeline endpoints.
type PipelineHandler struct {
	svc ports.PipelineService
}

// NewPipelineHandler constructs a PipelineHandler.
func NewPipelineHandler(svc ports.PipelineService) *PipelineHandler {
	return &PipelineHandler{svc: svc}
}

func (h *PipelineHandler) Create(c *gin.Context) {
	var req struct {
		Name          string                `json:"name" binding:"required"`
		RepositoryURL string                `json:"repository_url" binding:"required"`
		Branch        string                `json:"branch"`
		WebhookSecret string                `json:"webhook_secret"`
		Config        domain.PipelineConfig `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	pipeline, err := h.svc.CreatePipeline(c.Request.Context(), ports.CreatePipelineOptions{
		Name:          req.Name,
		RepositoryURL: req.RepositoryURL,
		Branch:        req.Branch,
		WebhookSecret: req.WebhookSecret,
		Config:        req.Config,
	})
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, pipeline)
}

func (h *PipelineHandler) List(c *gin.Context) {
	pipelines, err := h.svc.ListPipelines(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, pipelines)
}

func (h *PipelineHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	pipeline, err := h.svc.GetPipeline(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, pipeline)
}

func (h *PipelineHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	var req struct {
		Name          *string                `json:"name"`
		Branch        *string                `json:"branch"`
		WebhookSecret *string                `json:"webhook_secret"`
		Config        *domain.PipelineConfig `json:"config"`
		Status        *domain.PipelineStatus `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	pipeline, err := h.svc.UpdatePipeline(c.Request.Context(), id, ports.UpdatePipelineOptions{
		Name:          req.Name,
		Branch:        req.Branch,
		WebhookSecret: req.WebhookSecret,
		Config:        req.Config,
		Status:        req.Status,
	})
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, pipeline)
}

func (h *PipelineHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.DeletePipeline(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusNoContent, nil)
}

func (h *PipelineHandler) Trigger(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	var req struct {
		CommitHash  string                  `json:"commit_hash"`
		TriggerType domain.BuildTriggerType `json:"trigger_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	build, err := h.svc.TriggerBuild(c.Request.Context(), id, ports.TriggerBuildOptions{
		CommitHash:  req.CommitHash,
		TriggerType: req.TriggerType,
	})
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, build)
}

func (h *PipelineHandler) ListRuns(c *gin.Context) {
	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	builds, err := h.svc.ListBuildsByPipeline(c.Request.Context(), pipelineID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, builds)
}

func (h *PipelineHandler) GetRun(c *gin.Context) {
	buildID, err := uuid.Parse(c.Param("buildID"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	build, err := h.svc.GetBuild(c.Request.Context(), buildID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, build)
}

func (h *PipelineHandler) ListRunSteps(c *gin.Context) {
	buildID, err := uuid.Parse(c.Param("buildID"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	steps, err := h.svc.ListBuildSteps(c.Request.Context(), buildID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, steps)
}

func (h *PipelineHandler) ListRunLogs(c *gin.Context) {
	buildID, err := uuid.Parse(c.Param("buildID"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	limit := 200
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, parseErr := strconv.Atoi(rawLimit)
		if parseErr != nil {
			httputil.Error(c, fmt.Errorf("invalid limit: %w", parseErr))
			return
		}
		limit = parsed
	}

	logs, err := h.svc.ListBuildLogs(c.Request.Context(), buildID, limit)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, logs)
}

func (h *PipelineHandler) WebhookTrigger(c *gin.Context) {
	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httputil.Error(c, fmt.Errorf("failed to read payload: %w", err))
		return
	}

	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	event := c.GetHeader("X-GitHub-Event")
	signature := c.GetHeader("X-Hub-Signature-256")
	deliveryID := c.GetHeader("X-GitHub-Delivery")
	if provider == "gitlab" {
		event = c.GetHeader("X-Gitlab-Event")
		signature = c.GetHeader("X-Gitlab-Token")
		deliveryID = c.GetHeader("X-Gitlab-Event-UUID")
	}

	build, err := h.svc.TriggerBuildWebhook(c.Request.Context(), ports.WebhookTriggerOptions{
		PipelineID: pipelineID,
		Provider:   provider,
		Event:      event,
		Signature:  signature,
		DeliveryID: deliveryID,
		Payload:    payload,
	})
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if build == nil {
		httputil.Success(c, http.StatusAccepted, gin.H{"status": "ignored"})
		return
	}

	httputil.Success(c, http.StatusAccepted, build)
}
