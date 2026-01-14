package httphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockImageService struct {
	mock.Mock
}

func (m *mockImageService) RegisterImage(ctx context.Context, name, description, os, version string, isPublic bool) (*domain.Image, error) {
	args := m.Called(ctx, name, description, os, version, isPublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Image), args.Error(1)
}

func (m *mockImageService) UploadImage(ctx context.Context, id uuid.UUID, reader io.Reader) error {
	args := m.Called(ctx, id, reader)
	return args.Error(0)
}

func (m *mockImageService) GetImage(ctx context.Context, id uuid.UUID) (*domain.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Image), args.Error(1)
}

func (m *mockImageService) ListImages(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*domain.Image, error) {
	args := m.Called(ctx, userID, includePublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Image), args.Error(1)
}

func (m *mockImageService) DeleteImage(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestImageHandler_RegisterImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		expectedImage := &domain.Image{
			ID:   uuid.New(),
			Name: "test-image",
		}

		svc.On("RegisterImage", mock.Anything, "test-image", "desc", "linux", "1.0", true).Return(expectedImage, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/images", nil)

		body := map[string]interface{}{
			"name":        "test-image",
			"description": "desc",
			"os":          "linux",
			"version":     "1.0",
			"is_public":   true,
		}
		bodyBytes, _ := json.Marshal(body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		handler.RegisterImage(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("invalid input", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/images", nil)
		c.Request.Body = io.NopCloser(bytes.NewBufferString(`{}`)) // Missing required fields

		handler.RegisterImage(c)

		assert.NotEqual(t, http.StatusCreated, w.Code)
	})
}

func TestImageHandler_UploadImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		imageID := uuid.New()

		svc.On("UploadImage", mock.Anything, imageID, mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "image.qcow2")
		_, err := part.Write([]byte("image content"))
		assert.NoError(t, err)
		err = writer.Close()
		assert.NoError(t, err)

		c.Request = httptest.NewRequest("POST", "/api/v1/images/"+imageID.String()+"/upload", body)
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())
		c.Params = []gin.Param{{Key: "id", Value: imageID.String()}}

		handler.UploadImage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/images/invalid/upload", nil)
		c.Params = []gin.Param{{Key: "id", Value: "invalid"}}

		handler.UploadImage(c)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestImageHandler_ListImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		userID := uuid.New()
		images := []*domain.Image{{ID: uuid.New(), Name: "img1"}}

		svc.On("ListImages", mock.Anything, userID, true).Return(images, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/images", nil)
		c.Set("userID", userID)

		handler.ListImages(c)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestImageHandler_GetImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		imageID := uuid.New()
		img := &domain.Image{ID: imageID, Name: "img1"}

		svc.On("GetImage", mock.Anything, imageID).Return(img, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/images/"+imageID.String(), nil)
		c.Params = []gin.Param{{Key: "id", Value: imageID.String()}}

		handler.GetImage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/images/invalid", nil)
		c.Params = []gin.Param{{Key: "id", Value: "invalid"}}

		handler.GetImage(c)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestImageHandler_DeleteImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		imageID := uuid.New()

		svc.On("DeleteImage", mock.Anything, imageID).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/api/v1/images/"+imageID.String(), nil)
		c.Params = []gin.Param{{Key: "id", Value: imageID.String()}}

		handler.DeleteImage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}
