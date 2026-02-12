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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockElasticIPService struct {
	mock.Mock
}

const eipRoute = "/elastic-ips"

func (m *mockElasticIPService) AllocateIP(ctx context.Context) (*domain.ElasticIP, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ElasticIP), args.Error(1)
}

func (m *mockElasticIPService) ReleaseIP(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockElasticIPService) AssociateIP(ctx context.Context, id uuid.UUID, instanceID uuid.UUID) (*domain.ElasticIP, error) {
	args := m.Called(ctx, id, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ElasticIP), args.Error(1)
}

func (m *mockElasticIPService) DisassociateIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ElasticIP), args.Error(1)
}

func (m *mockElasticIPService) ListElasticIPs(ctx context.Context) ([]*domain.ElasticIP, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.ElasticIP), args.Error(1)
}

func (m *mockElasticIPService) GetElasticIP(ctx context.Context, id uuid.UUID) (*domain.ElasticIP, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ElasticIP), args.Error(1)
}

func setupElasticIPHandlerTest() (*mockElasticIPService, *ElasticIPHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockElasticIPService)
	handler := NewElasticIPHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestElasticIPHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		setupMock      func(svc *mockElasticIPService, eipID, instID uuid.UUID)
		body           interface{}
		expectedStatus int
	}{
		{
			name:   "Allocate",
			method: "POST",
			url:    eipRoute,
			setupMock: func(svc *mockElasticIPService, eipID, instID uuid.UUID) {
				svc.On("AllocateIP", mock.Anything).Return(&domain.ElasticIP{ID: eipID, PublicIP: "1.2.3.4"}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "List",
			method: "GET",
			url:    eipRoute,
			setupMock: func(svc *mockElasticIPService, eipID, instID uuid.UUID) {
				svc.On("ListElasticIPs", mock.Anything).Return([]*domain.ElasticIP{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Associate",
			method: "POST",
			url:    "/elastic-ips/{id}/associate",
			setupMock: func(svc *mockElasticIPService, eipID, instID uuid.UUID) {
				svc.On("AssociateIP", mock.Anything, eipID, instID).Return(&domain.ElasticIP{}, nil).Once()
			},
			body:           map[string]string{"instance_id": ""}, // Filled in loop
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, handler, r := setupElasticIPHandlerTest()
			r.POST(eipRoute, handler.Allocate)
			r.GET(eipRoute, handler.List)
			r.POST("/elastic-ips/:id/associate", handler.Associate)

			eipID := uuid.New()
			instID := uuid.New()
			tt.setupMock(svc, eipID, instID)

			url := tt.url
			if url == "/elastic-ips/{id}/associate" {
				url = "/elastic-ips/" + eipID.String() + "/associate"
			}

			var body []byte
			if tt.name == "Associate" {
				body, _ = json.Marshal(map[string]string{"instance_id": instID.String()})
			}

			w := httptest.NewRecorder()
			req, err := http.NewRequest(tt.method, url, bytes.NewBuffer(body))
			require.NoError(t, err)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}
