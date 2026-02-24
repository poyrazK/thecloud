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
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testSubnetName    = "test-subnet"
	vpcsPathPrefix    = "/vpcs/"
	subnetsPathPrefix = "/subnets/"
	subnetsPath       = "/subnets"
)

type mockSubnetService struct {
	mock.Mock
}

func (m *mockSubnetService) CreateSubnet(ctx context.Context, vpcID uuid.UUID, name, cidrBlock, az string) (*domain.Subnet, error) {
	args := m.Called(ctx, vpcID, name, cidrBlock, az)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Subnet)
	return r0, args.Error(1)
}

func (m *mockSubnetService) ListSubnets(ctx context.Context, vpcID uuid.UUID) ([]*domain.Subnet, error) {
	args := m.Called(ctx, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).([]*domain.Subnet)
	return r0, args.Error(1)
}

func (m *mockSubnetService) GetSubnet(ctx context.Context, idOrName string, vpcID uuid.UUID) (*domain.Subnet, error) {
	args := m.Called(ctx, idOrName, vpcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	r0, _ := args.Get(0).(*domain.Subnet)
	return r0, args.Error(1)
}

func (m *mockSubnetService) DeleteSubnet(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestSubnetHandlerCreateSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	vpcID := uuid.New()
	subnetID := uuid.New()

	expectedSubnet := &domain.Subnet{
		ID:        subnetID,
		VPCID:     vpcID,
		Name:      testSubnetName,
		CIDRBlock: testutil.TestSubnetCIDR,
	}

	svc.On("CreateSubnet", mock.Anything, vpcID, testSubnetName, testutil.TestSubnetCIDR, "us-east-1a").Return(expectedSubnet, nil)

	reqBody := map[string]string{
		"name":              testSubnetName,
		"cidr_block":        testutil.TestSubnetCIDR,
		"availability_zone": "us-east-1a",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", vpcsPathPrefix+vpcID.String()+subnetsPath, bytes.NewBuffer(body))
	c.Request.Header.Set(contentType, applicationJSON)
	c.Params = gin.Params{{Key: "vpc_id", Value: vpcID.String()}}

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerCreateInvalidVpcID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", vpcsPathPrefix+"invalid-uuid/subnets", nil)
	c.Params = gin.Params{{Key: "vpc_id", Value: "invalid-uuid"}}

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerCreateInvalidRequest(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	vpcID := uuid.New()

	reqBody := map[string]string{
		"name": "missing-cidr",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", vpcsPathPrefix+vpcID.String()+subnetsPath, bytes.NewBuffer(body))
	c.Request.Header.Set(contentType, applicationJSON)
	c.Params = gin.Params{{Key: "vpc_id", Value: vpcID.String()}}

	handler.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerCreateServiceError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	vpcID := uuid.New()

	svc.On("CreateSubnet", mock.Anything, vpcID, testSubnetName, testutil.TestSubnetCIDR, mock.Anything).Return(nil, assert.AnError)

	reqBody := map[string]string{
		"name":       testSubnetName,
		"cidr_block": testutil.TestSubnetCIDR,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", vpcsPathPrefix+vpcID.String()+subnetsPath, bytes.NewBuffer(body))
	c.Request.Header.Set(contentType, applicationJSON)
	c.Params = gin.Params{{Key: "vpc_id", Value: vpcID.String()}}

	handler.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerListSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	vpcID := uuid.New()

	expectedSubnets := []*domain.Subnet{
		{ID: uuid.New(), Name: "subnet-1"},
		{ID: uuid.New(), Name: "subnet-2"},
	}

	svc.On("ListSubnets", mock.Anything, vpcID).Return(expectedSubnets, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", vpcsPathPrefix+vpcID.String()+subnetsPath, nil)
	c.Params = gin.Params{{Key: "vpc_id", Value: vpcID.String()}}

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerListInvalidVpcID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", vpcsPathPrefix+"invalid/subnets", nil)
	c.Params = gin.Params{{Key: "vpc_id", Value: "invalid"}}

	handler.List(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerListServiceError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	vpcID := uuid.New()

	svc.On("ListSubnets", mock.Anything, vpcID).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", vpcsPathPrefix+vpcID.String()+subnetsPath, nil)
	c.Params = gin.Params{{Key: "vpc_id", Value: vpcID.String()}}

	handler.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerGetSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	subnetID := uuid.New()

	expectedSubnet := &domain.Subnet{
		ID:   subnetID,
		Name: testSubnetName,
	}

	svc.On("GetSubnet", mock.Anything, subnetID.String(), uuid.Nil).Return(expectedSubnet, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", subnetsPathPrefix+subnetID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: subnetID.String()}}

	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerGetServiceError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	subnetID := uuid.New()

	svc.On("GetSubnet", mock.Anything, subnetID.String(), uuid.Nil).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", subnetsPathPrefix+subnetID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: subnetID.String()}}

	handler.Get(c)

	assert.NotEqual(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerDeleteSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	subnetID := uuid.New()

	svc.On("DeleteSubnet", mock.Anything, subnetID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", subnetsPathPrefix+subnetID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: subnetID.String()}}

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerDeleteInvalidID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", subnetsPathPrefix+"invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertExpectations(t)
}

func TestSubnetHandlerDeleteServiceError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := new(mockSubnetService)
	handler := NewSubnetHandler(svc)

	subnetID := uuid.New()

	svc.On("DeleteSubnet", mock.Anything, subnetID).Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("DELETE", subnetsPathPrefix+subnetID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: subnetID.String()}}

	handler.Delete(c)

	assert.NotEqual(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}
