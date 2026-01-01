package httphandlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/poyraz/cloud/internal/core/ports"
	"github.com/poyraz/cloud/pkg/httputil"
)

type EventHandler struct {
	svc ports.EventService
}

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
