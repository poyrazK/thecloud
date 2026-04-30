package httphandlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type dashboardServiceMock struct {
	mock.Mock
}

func (m *dashboardServiceMock) GetSummary(ctx context.Context) (*domain.ResourceSummary, error) {
	args := m.Called(ctx)
	// Helper for checking return value
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.ResourceSummary)
	return r0, args.Error(1)
}

func (m *dashboardServiceMock) GetRecentEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Event)
	return r0, args.Error(1)
}

func (m *dashboardServiceMock) GetStats(ctx context.Context) (*domain.DashboardStats, error) {
	args := m.Called(ctx)
	// Check for nil return
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.DashboardStats)
	return r0, args.Error(1)
}

func setupDashboardHandlerTest(_ *testing.T) (*dashboardServiceMock, *DashboardHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(dashboardServiceMock)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewDashboardHandler(mockSvc, logger, "*") // "*" for test allowlist
	r := gin.New()
	return mockSvc, handler, r
}

func TestDashboardHandlerGetSummary(t *testing.T) {
	t.Parallel()
	mockSvc, handler, r := setupDashboardHandlerTest(t)
	defer mockSvc.AssertExpectations(t)

	r.GET("/summary", handler.GetSummary)

	summary := &domain.ResourceSummary{TotalInstances: 5}
	mockSvc.On("GetSummary", mock.Anything).Return(summary, nil)

	req, err := http.NewRequest("GET", "/summary", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var wrapper struct {
		Data domain.ResourceSummary `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &wrapper))
	assert.Equal(t, 5, wrapper.Data.TotalInstances)
}

func TestDashboardHandlerGetRecentEvents(t *testing.T) {
	t.Parallel()
	mockSvc, handler, r := setupDashboardHandlerTest(t)
	defer mockSvc.AssertExpectations(t)

	r.GET("/events", handler.GetRecentEvents)

	events := []*domain.Event{{ID: uuid.New(), Action: "TEST"}}
	mockSvc.On("GetRecentEvents", mock.Anything, 10).Return(events, nil)

	req, err := http.NewRequest("GET", "/events?limit=10", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var wrapper struct {
		Data []*domain.Event `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &wrapper))
	assert.Len(t, wrapper.Data, 1)
}

func TestDashboardHandlerGetStats(t *testing.T) {
	t.Parallel()
	mockSvc, handler, r := setupDashboardHandlerTest(t)
	defer mockSvc.AssertExpectations(t)

	r.GET("/stats", handler.GetStats)

	stats := &domain.DashboardStats{
		CPUHistory: []domain.MetricPoint{{Value: 10.1}},
	}
	mockSvc.On("GetStats", mock.Anything).Return(stats, nil)

	req, err := http.NewRequest("GET", "/stats", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var wrapper struct {
		Data domain.DashboardStats `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &wrapper))
	assert.Len(t, wrapper.Data.CPUHistory, 1)
	assert.InDelta(t, 10.1, wrapper.Data.CPUHistory[0].Value, 0.01)
}

func TestDashboardHandlerStreamEvents(t *testing.T) {
	t.Parallel()
	mockSvc, handler, r := setupDashboardHandlerTest(t)
	defer mockSvc.AssertExpectations(t)

	r.GET("/stream", handler.StreamEvents)

	summary := &domain.ResourceSummary{TotalInstances: 10}
	mockSvc.On("GetSummary", mock.Anything).Return(summary, nil)

	req, err := http.NewRequest("GET", "/stream", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	go r.ServeHTTP(w, req)

	time.Sleep(100 * time.Millisecond)
	cancel()

	assert.Contains(t, w.Header().Get("Content-Type"), "text/event-stream")
	assert.Contains(t, w.Body.String(), "event:summary")
}

func TestDashboardHandlerGetRecentEventsLimits(t *testing.T) {
	t.Parallel()
	mockSvc, handler, r := setupDashboardHandlerTest(t)
	r.GET("/events", handler.GetRecentEvents)

	t.Run("Default", func(t *testing.T) {
		mockSvc.On("GetRecentEvents", mock.Anything, 10).Return([]*domain.Event{}, nil).Once()
		req, _ := http.NewRequest("GET", "/events", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid", func(t *testing.T) {
		mockSvc.On("GetRecentEvents", mock.Anything, 10).Return([]*domain.Event{}, nil).Once()
		req, _ := http.NewRequest("GET", "/events?limit=abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Cap", func(t *testing.T) {
		mockSvc.On("GetRecentEvents", mock.Anything, 100).Return([]*domain.Event{}, nil).Once()
		req, _ := http.NewRequest("GET", "/events?limit=200", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestDashboardHandlerErrors(t *testing.T) {
	t.Parallel()
	mockSvc, handler, r := setupDashboardHandlerTest(t)
	r.GET("/summary", handler.GetSummary)
	r.GET("/events", handler.GetRecentEvents)
	r.GET("/stats", handler.GetStats)

	t.Run("SummaryError", func(t *testing.T) {
		mockSvc.On("GetSummary", mock.Anything).Return(nil, assert.AnError)
		req, _ := http.NewRequest("GET", "/summary", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("EventsError", func(t *testing.T) {
		mockSvc.On("GetRecentEvents", mock.Anything, 10).Return(nil, assert.AnError)
		req, _ := http.NewRequest("GET", "/events", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("StatsError", func(t *testing.T) {
		mockSvc.On("GetStats", mock.Anything).Return(nil, assert.AnError)
		req, _ := http.NewRequest("GET", "/stats", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDashboardHandlerStreamEvents_OriginNotAllowed(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	mockSvc := new(dashboardServiceMock)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Only "https://allowed.example.com" is allowed — anything else is rejected
	handler := NewDashboardHandler(mockSvc, logger, "https://allowed.example.com")
	r := gin.New()
	r.GET("/stream", handler.StreamEvents)

	summary := &domain.ResourceSummary{TotalInstances: 10}
	mockSvc.On("GetSummary", mock.Anything).Return(summary, nil)

	req, _ := http.NewRequest("GET", "/stream", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	w := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	go r.ServeHTTP(w, req)
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Request should be rejected with 403 before SSE headers are sent
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDashboardHandlerStreamEvents_OriginAllowed(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	mockSvc := new(dashboardServiceMock)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewDashboardHandler(mockSvc, logger, "https://app.example.com")
	r := gin.New()
	r.GET("/stream", handler.StreamEvents)

	summary := &domain.ResourceSummary{TotalInstances: 10}
	mockSvc.On("GetSummary", mock.Anything).Return(summary, nil)

	req, _ := http.NewRequest("GET", "/stream", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	go r.ServeHTTP(w, req)
	time.Sleep(100 * time.Millisecond)
	cancel()

	assert.Contains(t, w.Header().Get("Content-Type"), "text/event-stream")
	assert.Contains(t, w.Body.String(), "event:summary")
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestDashboardHandlerStreamEvents_WildcardAllowsAny(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	mockSvc := new(dashboardServiceMock)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewDashboardHandler(mockSvc, logger, "*")
	r := gin.New()
	r.GET("/stream", handler.StreamEvents)

	summary := &domain.ResourceSummary{TotalInstances: 10}
	mockSvc.On("GetSummary", mock.Anything).Return(summary, nil)

	req, _ := http.NewRequest("GET", "/stream", nil)
	req.Header.Set("Origin", "https://anything.com")
	w := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	go r.ServeHTTP(w, req)
	time.Sleep(100 * time.Millisecond)
	cancel()

	assert.Contains(t, w.Header().Get("Content-Type"), "text/event-stream")
	assert.Contains(t, w.Body.String(), "event:summary")
}

func TestDashboardHandlerStreamEvents_EmptyAllowlistDenied(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	mockSvc := new(dashboardServiceMock)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Empty allowlist = fail-closed
	handler := NewDashboardHandler(mockSvc, logger)
	r := gin.New()
	r.GET("/stream", handler.StreamEvents)

	req, _ := http.NewRequest("GET", "/stream", nil)
	req.Header.Set("Origin", "https://anything.com")
	w := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	go r.ServeHTTP(w, req)
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Empty allowlist → deny all cross-origin → 403
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDashboardHandlerStreamEvents_SameOriginNoHeader(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	mockSvc := new(dashboardServiceMock)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// With wildcard, same-origin (no Origin header) should be allowed
	handler := NewDashboardHandler(mockSvc, logger, "*")
	r := gin.New()
	r.GET("/stream", handler.StreamEvents)

	summary := &domain.ResourceSummary{TotalInstances: 10}
	mockSvc.On("GetSummary", mock.Anything).Return(summary, nil)

	req, _ := http.NewRequest("GET", "/stream", nil)
	// No Origin header set — same-origin request
	w := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	go r.ServeHTTP(w, req)
	time.Sleep(100 * time.Millisecond)
	cancel()

	// No Origin header = same-origin → allowed with wildcard
	assert.Contains(t, w.Header().Get("Content-Type"), "text/event-stream")
	assert.Contains(t, w.Body.String(), "event:summary")
}
