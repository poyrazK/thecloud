package httphandlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// LogHandler handles log-related requests.
type LogHandler struct {
	svc ports.LogService
}

// NewLogHandler creates a new LogHandler.
func NewLogHandler(svc ports.LogService) *LogHandler {
	return &LogHandler{svc: svc}
}

// Search searches logs with filtering.
// @Summary Search logs
// @Description Search and filter resource logs
// @Tags logs
// @Produce json
// @Security APIKeyAuth
// @Param resource_id query string false "Resource ID"
// @Param resource_type query string false "Resource Type"
// @Param level query string false "Log Level"
// @Param search query string false "Search Keyword"
// @Param start_time query string false "Start Time (RFC3339)"
// @Param end_time query string false "End Time (RFC3339)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} httputil.Response
// @Router /logs [get]
func (h *LogHandler) Search(c *gin.Context) {
	tenantID := appcontext.TenantIDFromContext(c.Request.Context())
	
	query := domain.LogQuery{
		TenantID:     tenantID,
		ResourceID:   c.Query("resource_id"),
		ResourceType: c.Query("resource_type"),
		Level:        c.Query("level"),
		Search:       c.Query("search"),
	}

	if st := c.Query("start_time"); st != "" {
		t, err := time.Parse(time.RFC3339, st)
		if err != nil {
			httputil.Error(c, errors.New(errors.InvalidInput, "invalid start_time format; expected RFC3339"))
			return
		}
		query.StartTime = &t
	}
	if et := c.Query("end_time"); et != "" {
		t, err := time.Parse(time.RFC3339, et)
		if err != nil {
			httputil.Error(c, errors.New(errors.InvalidInput, "invalid end_time format; expected RFC3339"))
			return
		}
		query.EndTime = &t
	}

	// Pagination
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid limit; must be a number"))
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid offset; must be a number"))
		return
	}

	query.Limit = limit
	query.Offset = offset

	entries, total, err := h.svc.SearchLogs(c.Request.Context(), query)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{
		"entries": entries,
		"total":   total,
		"limit":   query.Limit,
		"offset":  query.Offset,
	})
}

// GetByResource returns logs for a specific resource.
// @Summary Get logs by resource
// @Description Gets logs for a specific resource ID
// @Tags logs
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Resource ID"
// @Param limit query int false "Limit"
// @Success 200 {object} httputil.Response
// @Router /logs/{id} [get]
func (h *LogHandler) GetByResource(c *gin.Context) {
	id := c.Param("id")
	tenantID := appcontext.TenantIDFromContext(c.Request.Context())

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	query := domain.LogQuery{
		TenantID:   tenantID,
		ResourceID: id,
		Limit:      limit,
	}

	entries, total, err := h.svc.SearchLogs(c.Request.Context(), query)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{
		"entries": entries,
		"total":   total,
	})
}
