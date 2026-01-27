// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

const invalidJobIDMsg = "Invalid job ID"

// CronHandler handles scheduled job HTTP endpoints.
type CronHandler struct {
	svc ports.CronService
}

// NewCronHandler constructs a CronHandler.
func NewCronHandler(svc ports.CronService) *CronHandler {
	return &CronHandler{svc: svc}
}

func (h *CronHandler) CreateJob(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		Schedule      string `json:"schedule" binding:"required"`
		TargetURL     string `json:"target_url" binding:"required"`
		TargetMethod  string `json:"target_method"`
		TargetPayload string `json:"target_payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "Invalid request body"))
		return
	}

	if req.TargetMethod == "" {
		req.TargetMethod = "POST"
	}

	job, err := h.svc.CreateJob(c.Request.Context(), req.Name, req.Schedule, req.TargetURL, req.TargetMethod, req.TargetPayload)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, job)
}

func (h *CronHandler) ListJobs(c *gin.Context) {
	jobs, err := h.svc.ListJobs(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, jobs)
}

func (h *CronHandler) GetJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidJobIDMsg))
		return
	}

	job, err := h.svc.GetJob(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, job)
}

func (h *CronHandler) PauseJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidJobIDMsg))
		return
	}

	if err := h.svc.PauseJob(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "Job paused"})
}

func (h *CronHandler) ResumeJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidJobIDMsg))
		return
	}

	if err := h.svc.ResumeJob(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "Job resumed"})
}

func (h *CronHandler) DeleteJob(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidJobIDMsg))
		return
	}

	if err := h.svc.DeleteJob(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "Job deleted"})
}
