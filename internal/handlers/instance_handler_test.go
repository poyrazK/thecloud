package httphandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyraz/cloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type instanceServiceMock struct {
	mock.Mock
}

func (m *instanceServiceMock) LaunchInstance(ctx context.Context, name, image, ports string, vpcID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error) {
	args := m.Called(ctx, name, image, ports, vpcID, volumes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (m *instanceServiceMock) StopInstance(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func (m *instanceServiceMock) ListInstances(ctx context.Context) ([]*domain.Instance, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.Instance), args.Error(1)
}

func (m *instanceServiceMock) GetInstance(ctx context.Context, idOrName string) (*domain.Instance, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Instance), args.Error(1)
}

func (m *instanceServiceMock) GetInstanceLogs(ctx context.Context, idOrName string) (string, error) {
	args := m.Called(ctx, idOrName)
	return args.String(0), args.Error(1)
}

func (m *instanceServiceMock) GetInstanceStats(ctx context.Context, idOrName string) (*domain.InstanceStats, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InstanceStats), args.Error(1)
}

func (m *instanceServiceMock) TerminateInstance(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func TestInstanceHandler_LaunchRejectsEmptyImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(instanceServiceMock)
	handler := NewInstanceHandler(mockSvc)
	r := gin.New()
	r.POST("/instances", handler.Launch)

	body := `{"name":"test-inst","image":"   "}`
	req := httptest.NewRequest(http.MethodPost, "/instances", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var wrapper struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &wrapper))
	assert.Equal(t, "INVALID_INPUT", wrapper.Error.Type)
	mockSvc.AssertNotCalled(t, "LaunchInstance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
