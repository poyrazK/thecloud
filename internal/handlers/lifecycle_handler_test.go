package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockLifecycleService struct {
	mock.Mock
}

func (m *mockLifecycleService) CreateRule(ctx context.Context, bucket string, prefix string, expirationDays int, enabled bool) (*domain.LifecycleRule, error) {
	args := m.Called(ctx, bucket, prefix, expirationDays, enabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LifecycleRule), args.Error(1)
}

func (m *mockLifecycleService) ListRules(ctx context.Context, bucket string) ([]*domain.LifecycleRule, error) {
	args := m.Called(ctx, bucket)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LifecycleRule), args.Error(1)
}

func (m *mockLifecycleService) DeleteRule(ctx context.Context, bucket string, ruleID string) error {
	args := m.Called(ctx, bucket, ruleID)
	return args.Error(0)
}

const (
	testBucketName    = "test-bucket"
	testLifecyclePath = "/storage/buckets/test-bucket/lifecycle"
	testPrefix        = "logs/"
)

func setupLifecycleHandlerTest(t *testing.T) (*mockLifecycleService, *LifecycleHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockLifecycleService)

	// Automatically assert mock expectations after test cleanup
	t.Cleanup(func() {
		svc.AssertExpectations(t)
	})

	handler := NewLifecycleHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestLifecycleHandlerCreateRule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(*mockLifecycleService)
		expectedStatus int
	}{
		{
			name: "Success",
			body: map[string]interface{}{
				"prefix":          testPrefix,
				"expiration_days": 30,
				"enabled":         true,
			},
			setupMock: func(m *mockLifecycleService) {
				rule := &domain.LifecycleRule{ID: uuid.New(), BucketName: testBucketName, Prefix: testPrefix, ExpirationDays: 30, Enabled: true}
				m.On("CreateRule", mock.Anything, testBucketName, testPrefix, 30, true).Return(rule, nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid Input",
			body: map[string]interface{}{},
			setupMock: func(m *mockLifecycleService) {
				// No calls expected
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			body: map[string]interface{}{
				"prefix":          testPrefix,
				"expiration_days": 30,
				"enabled":         true,
			},
			setupMock: func(m *mockLifecycleService) {
				m.On("CreateRule", mock.Anything, testBucketName, testPrefix, 30, true).Return(nil, errors.New(errors.Internal, "error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, handler, r := setupLifecycleHandlerTest(t)
			r.POST("/storage/buckets/:bucket/lifecycle", handler.CreateRule)

			tc.setupMock(svc)

			bodyBytes, _ := json.Marshal(tc.body)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", testLifecyclePath, bytes.NewBuffer(bodyBytes))
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestLifecycleHandlerListRules(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupLifecycleHandlerTest(t)
	r.GET("/storage/buckets/:bucket/lifecycle", handler.ListRules)

	t.Run("Success", func(t *testing.T) {
		svc.On("ListRules", mock.Anything, testBucketName).Return([]*domain.LifecycleRule{}, nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testLifecyclePath, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		svc.On("ListRules", mock.Anything, testBucketName).Return(nil, errors.New(errors.Internal, "error")).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", testLifecyclePath, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestLifecycleHandlerDeleteRule(t *testing.T) {
	t.Parallel()
	svc, handler, r := setupLifecycleHandlerTest(t)
	r.DELETE("/storage/buckets/:bucket/lifecycle/:id", handler.DeleteRule)

	ruleID := uuid.New().String()

	t.Run("Success", func(t *testing.T) {
		svc.On("DeleteRule", mock.Anything, testBucketName, ruleID).Return(nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", testLifecyclePath+"/"+ruleID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		svc.On("DeleteRule", mock.Anything, testBucketName, ruleID).Return(errors.New(errors.NotFound, "not found")).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", testLifecyclePath+"/"+ruleID, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
