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
)

type mockVpcService struct {
	mock.Mock
}

func (m *mockVpcService) CreateVPC(ctx context.Context, name, cidrBlock string) (*domain.VPC, error) {
	args := m.Called(ctx, name, cidrBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}

func (m *mockVpcService) ListVPCs(ctx context.Context) ([]*domain.VPC, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.VPC), args.Error(1)
}

func (m *mockVpcService) GetVPC(ctx context.Context, idOrName string) (*domain.VPC, error) {
	args := m.Called(ctx, idOrName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VPC), args.Error(1)
}

func (m *mockVpcService) DeleteVPC(ctx context.Context, idOrName string) error {
	args := m.Called(ctx, idOrName)
	return args.Error(0)
}

func setupVpcHandlerTest(t *testing.T) (*mockVpcService, *VpcHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockVpcService)
	handler := NewVpcHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestVpcHandler_Create(t *testing.T) {
	svc, handler, r := setupVpcHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.POST("/vpcs", handler.Create)

	vpc := &domain.VPC{ID: uuid.New(), Name: "test-vpc"}
	svc.On("CreateVPC", mock.Anything, "test-vpc", "10.0.0.0/16").Return(vpc, nil)

	body, err := json.Marshal(map[string]string{"name": "test-vpc", "cidr_block": "10.0.0.0/16"})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/vpcs", bytes.NewBuffer(body))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestVpcHandler_List(t *testing.T) {
	svc, handler, r := setupVpcHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/vpcs", handler.List)

	vpcs := []*domain.VPC{{ID: uuid.New(), Name: "vpc1"}}
	svc.On("ListVPCs", mock.Anything).Return(vpcs, nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/vpcs", nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVpcHandler_Get(t *testing.T) {
	svc, handler, r := setupVpcHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/vpcs/:id", handler.Get)

	vpcID := uuid.New().String()
	vpc := &domain.VPC{ID: uuid.New(), Name: "vpc1"}
	svc.On("GetVPC", mock.Anything, vpcID).Return(vpc, nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/vpcs/"+vpcID, nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVpcHandler_Delete(t *testing.T) {
	svc, handler, r := setupVpcHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE("/vpcs/:id", handler.Delete)

	vpcID := uuid.New().String()
	svc.On("DeleteVPC", mock.Anything, vpcID).Return(nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/vpcs/"+vpcID, nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
