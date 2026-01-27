// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

const (
	invalidRequestBodyMsg = "Invalid request body"
	invalidTopicIDMsg     = "Invalid topic ID"
)

// NotifyHandler handles notification HTTP endpoints.
type NotifyHandler struct {
	svc ports.NotifyService
}

// NewNotifyHandler constructs a NotifyHandler.
func NewNotifyHandler(svc ports.NotifyService) *NotifyHandler {
	return &NotifyHandler{svc: svc}
}

func (h *NotifyHandler) CreateTopic(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidRequestBodyMsg))
		return
	}

	topic, err := h.svc.CreateTopic(c.Request.Context(), req.Name)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, topic)
}

func (h *NotifyHandler) ListTopics(c *gin.Context) {
	topics, err := h.svc.ListTopics(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, topics)
}

func (h *NotifyHandler) DeleteTopic(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidTopicIDMsg))
		return
	}

	if err := h.svc.DeleteTopic(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "Topic deleted"})
}

func (h *NotifyHandler) Subscribe(c *gin.Context) {
	topicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidTopicIDMsg))
		return
	}

	var req struct {
		Protocol domain.SubscriptionProtocol `json:"protocol" binding:"required"`
		Endpoint string                      `json:"endpoint" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidRequestBodyMsg))
		return
	}

	sub, err := h.svc.Subscribe(c.Request.Context(), topicID, req.Protocol, req.Endpoint)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, sub)
}

func (h *NotifyHandler) ListSubscriptions(c *gin.Context) {
	topicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidTopicIDMsg))
		return
	}

	subs, err := h.svc.ListSubscriptions(c.Request.Context(), topicID)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, subs)
}

func (h *NotifyHandler) Unsubscribe(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "Invalid subscription ID"))
		return
	}

	if err := h.svc.Unsubscribe(c.Request.Context(), id); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "Unsubscribed"})
}

func (h *NotifyHandler) Publish(c *gin.Context) {
	topicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "Invalid topic ID"))
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, invalidRequestBodyMsg))
		return
	}

	if err := h.svc.Publish(c.Request.Context(), topicID, req.Message); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"message": "Message published"})
}
