package httphandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockLogService struct {
	mock.Mock
}

func (m *mockLogService) IngestLogs(ctx context.Context, entries []*domain.LogEntry) error {
	return m.Called(ctx, entries).Error(0)
}
func (m *mockLogService) SearchLogs(ctx context.Context, query domain.LogQuery) ([]*domain.LogEntry, int, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	r0, _ := args.Get(0).([]*domain.LogEntry)
	return r0, args.Int(1), args.Error(2)
}
func (m *mockLogService) RunRetentionPolicy(ctx context.Context, days int) error {
	return m.Called(ctx, days).Error(0)
}

func setupLogHandlerTest() (*mockLogService, *LogHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockLogService)
	handler := NewLogHandler(mockSvc)
	r := gin.New()
	return mockSvc, handler, r
}

func TestLogHandlerSearch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc, handler, r := setupLogHandlerTest()
		r.GET("/logs", handler.Search)

		expectedLogs := []*domain.LogEntry{
			{Message: "log 1"},
		}

		mockSvc.On("SearchLogs", mock.Anything, mock.MatchedBy(func(q domain.LogQuery) bool {
			return q.ResourceID == "res-1"
		})).Return(expectedLogs, 1, nil)

		req := httptest.NewRequest(http.MethodGet, "/logs?resource_id=res-1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp struct {
			Data struct {
				Entries []domain.LogEntry `json:"entries"`
				Total   int               `json:"total"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Data.Total)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc, handler, r := setupLogHandlerTest()
		r.GET("/logs", handler.Search)

		mockSvc.On("SearchLogs", mock.Anything, mock.Anything).Return(nil, 0, errors.New("db fail"))

		req := httptest.NewRequest(http.MethodGet, "/logs", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("with time filters", func(t *testing.T) {
		mockSvc, handler, r := setupLogHandlerTest()
		r.GET("/logs", handler.Search)

		mockSvc.On("SearchLogs", mock.Anything, mock.MatchedBy(func(q domain.LogQuery) bool {
			return q.StartTime != nil && q.EndTime != nil
		})).Return([]*domain.LogEntry{}, 0, nil)

		req := httptest.NewRequest(http.MethodGet, "/logs?start_time=2026-01-01T00:00:00Z&end_time=2026-01-02T00:00:00Z", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestLogHandlerGetByResource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc, handler, r := setupLogHandlerTest()
		r.GET("/logs/:id", handler.GetByResource)

		id := uuid.New().String()
		expectedLogs := []*domain.LogEntry{
			{Message: "resource log"},
		}

		mockSvc.On("SearchLogs", mock.Anything, mock.MatchedBy(func(q domain.LogQuery) bool {
			return q.ResourceID == id
		})).Return(expectedLogs, 1, nil)

		req := httptest.NewRequest(http.MethodGet, "/logs/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockSvc, handler, r := setupLogHandlerTest()
		r.GET("/logs/:id", handler.GetByResource)

		mockSvc.On("SearchLogs", mock.Anything, mock.Anything).Return(nil, 0, errors.New("fail"))

		req := httptest.NewRequest(http.MethodGet, "/logs/res-1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
