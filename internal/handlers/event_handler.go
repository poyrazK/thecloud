// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// EventHandler handles event listing endpoints.
type EventHandler struct {
	svc ports.EventService
}

// NewEventHandler constructs an EventHandler.
func NewEventHandler(svc ports.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	events, err := h.svc.ListEvents(c.Request.Context(), limit)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, events)
}
