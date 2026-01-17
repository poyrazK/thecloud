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
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	imagesAPI        = "/api/v1/images"
	imageIDParam     = "id"
	testImage        = "test-image"
	errInvalidID     = "invalid id"
	imagePathInvalid = "/invalid"
	uploadSuffix     = "/upload"
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

func TestImageHandlerRegisterImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		expectedImage := &domain.Image{
			ID:   uuid.New(),
			Name: testImage,
		}

		svc.On("RegisterImage", mock.Anything, testImage, "desc", "linux", "1.0", true).Return(expectedImage, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", imagesAPI, nil)

		body := map[string]interface{}{
			"name":        testImage,
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
		c.Request = httptest.NewRequest("POST", imagesAPI, nil)
		c.Request.Body = io.NopCloser(bytes.NewBufferString(`{}`)) // Missing required fields

		handler.RegisterImage(c)

		assert.NotEqual(t, http.StatusCreated, w.Code)
	})
}

func TestImageHandlerUploadImage(t *testing.T) {
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

		c.Request = httptest.NewRequest("POST", imagesAPI+"/"+imageID.String()+uploadSuffix, body)
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())
		c.Params = []gin.Param{{Key: imageIDParam, Value: imageID.String()}}

		handler.UploadImage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run(errInvalidID, func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", imagesAPI+imagePathInvalid+uploadSuffix, nil)
		c.Params = []gin.Param{{Key: imageIDParam, Value: "invalid"}}

		handler.UploadImage(c)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestImageHandlerListImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		userID := uuid.New()
		images := []*domain.Image{{ID: uuid.New(), Name: "img1"}}

		svc.On("ListImages", mock.Anything, userID, true).Return(images, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", imagesAPI, nil)
		c.Set("userID", userID)

		handler.ListImages(c)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestImageHandlerGetImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		imageID := uuid.New()
		img := &domain.Image{ID: imageID, Name: "img1"}

		svc.On("GetImage", mock.Anything, imageID).Return(img, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", imagesAPI+"/"+imageID.String(), nil)
		c.Params = []gin.Param{{Key: imageIDParam, Value: imageID.String()}}

		handler.GetImage(c)

		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run(errInvalidID, func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", imagesAPI+"/invalid", nil)
		c.Params = []gin.Param{{Key: imageIDParam, Value: "invalid"}}

		handler.GetImage(c)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)
		id := uuid.New()
		svc.On("GetImage", mock.Anything, id).Return(nil, errors.New(errors.Internal, "error"))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Params = []gin.Param{{Key: imageIDParam, Value: id.String()}}
		handler.GetImage(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestImageHandlerDeleteImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)
		imageID := uuid.New()
		svc.On("DeleteImage", mock.Anything, imageID).Return(nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", imagesAPI+"/"+imageID.String(), nil)
		c.Params = []gin.Param{{Key: imageIDParam, Value: imageID.String()}}
		handler.DeleteImage(c)
		assert.Equal(t, http.StatusOK, w.Code)
		svc.AssertExpectations(t)
	})

	t.Run(errInvalidID, func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", imagesAPI+imagePathInvalid, nil)
		c.Params = []gin.Param{{Key: imageIDParam, Value: "invalid"}}

		handler.DeleteImage(c)

		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)

		imageID := uuid.New()
		svc.On("DeleteImage", mock.Anything, imageID).Return(errors.New(errors.Internal, "error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", imagesAPI+"/"+imageID.String(), nil)
		c.Params = []gin.Param{{Key: imageIDParam, Value: imageID.String()}}

		handler.DeleteImage(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		svc.AssertExpectations(t)
	})
}

func TestImageHandlerAdditionalErrors(t *testing.T) {
	t.Run("RegisterImageServiceError", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)
		svc.On("RegisterImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New(errors.Internal, "error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", imagesAPI, nil)
		bodyBytes, _ := json.Marshal(map[string]interface{}{"name": "n", "os": "o", "version": "v"})
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		handler.RegisterImage(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UploadImageInvalidID", func(t *testing.T) {
		_, handler, r := setupImageHandlerTest(t)
		r.POST("/images/:id"+uploadSuffix, handler.UploadImage)
		req, _ := http.NewRequest("POST", "/images/invalid/upload", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UploadImageNoFile", func(t *testing.T) {
		_, handler, _ := setupImageHandlerTest(t)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: imageIDParam, Value: uuid.New().String()}}
		c.Request = httptest.NewRequest("POST", "/upload", nil) // No multipart
		handler.UploadImage(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ListImagesServiceError", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)
		svc.On("ListImages", mock.Anything, mock.Anything, true).Return(nil, errors.New(errors.Internal, "error"))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", imagesAPI, nil)
		c.Set("userID", uuid.New())
		handler.ListImages(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UploadImageServiceError", func(t *testing.T) {
		svc := new(mockImageService)
		handler := NewImageHandler(svc)
		id := uuid.New()
		svc.On("UploadImage", mock.Anything, id, mock.Anything).Return(errors.New(errors.Internal, "error"))

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "image.qcow2")
		_, _ = part.Write([]byte("image content"))
		_ = writer.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", uploadSuffix, body)
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())
		c.Params = []gin.Param{{Key: imageIDParam, Value: id.String()}}

		handler.UploadImage(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func setupImageHandlerTest(_ *testing.T) (*mockImageService, *ImageHandler, *gin.Engine) {
	svc := new(mockImageService)
	handler := NewImageHandler(svc)
	r := gin.New()
	return svc, handler, r
}
