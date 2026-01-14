// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// QueueHandler handles queue HTTP endpoints.
type QueueHandler struct {
	svc ports.QueueService
}

// NewQueueHandler constructs a QueueHandler.
func NewQueueHandler(svc ports.QueueService) *QueueHandler {
	return &QueueHandler{
		svc: svc,
	}
}

func (h *QueueHandler) Create(c *gin.Context) {
	var req struct {
		Name              string `json:"name" binding:"required"`
		VisibilityTimeout *int   `json:"visibility_timeout"`
		RetentionDays     *int   `json:"retention_days"`
		MaxMessageSize    *int   `json:"max_message_size"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	opts := &ports.CreateQueueOptions{
		VisibilityTimeout: req.VisibilityTimeout,
		RetentionDays:     req.RetentionDays,
		MaxMessageSize:    req.MaxMessageSize,
	}

	q, err := h.svc.CreateQueue(c.Request.Context(), req.Name, opts)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, q)
}

func (h *QueueHandler) List(c *gin.Context) {
	queues, err := h.svc.ListQueues(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, queues)
}

func (h *QueueHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	q, err := h.svc.GetQueue(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, q)
}

func (h *QueueHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.DeleteQueue(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusNoContent, nil)
}

func (h *QueueHandler) SendMessage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, err)
		return
	}

	msg, err := h.svc.SendMessage(c.Request.Context(), id, req.Body)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusCreated, msg)
}

func (h *QueueHandler) ReceiveMessages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	maxStr := c.DefaultQuery("max_messages", "1")
	var max int
	if _, err := fmt.Sscanf(maxStr, "%d", &max); err != nil {
		max = 1
	}

	msgs, err := h.svc.ReceiveMessages(c.Request.Context(), id, max)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, msgs)
}

func (h *QueueHandler) DeleteMessage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	handle := c.Param("handle")
	if handle == "" {
		httputil.Error(c, fmt.Errorf("receipt handle is required"))
		return
	}

	if err := h.svc.DeleteMessage(c.Request.Context(), id, handle); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusNoContent, nil)
}

func (h *QueueHandler) Purge(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, err)
		return
	}

	if err := h.svc.PurgeQueue(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusNoContent, nil)
}
