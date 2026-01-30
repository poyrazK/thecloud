package httphandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/httputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type instanceTypeServiceMock struct {
	mock.Mock
}

func (m *instanceTypeServiceMock) List(ctx context.Context) ([]*domain.InstanceType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InstanceType), args.Error(1)
}

func TestInstanceTypeHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(instanceTypeServiceMock)
	handler := NewInstanceTypeHandler(mockSvc)
	r := gin.New()
	r.GET("/instance-types", handler.List)

	t.Run("success", func(t *testing.T) {
		expectedTypes := []*domain.InstanceType{
			{ID: "basic-1", Name: "Basic Small", VCPUs: 1, MemoryMB: 1024, DiskGB: 10},
		}
		mockSvc.On("List", mock.Anything).Return(expectedTypes, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/instance-types", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp httputil.Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Data)

		data, _ := json.Marshal(resp.Data)
		var actualTypes []*domain.InstanceType
		err = json.Unmarshal(data, &actualTypes)
		assert.NoError(t, err)
		assert.Len(t, actualTypes, 1)
		assert.Equal(t, "basic-1", actualTypes[0].ID)

		mockSvc.AssertExpectations(t)
	})
}
