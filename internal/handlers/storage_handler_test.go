package httphandlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testFileName      = "test.txt"
	storageBucketName = "b1"
	storageObjectPath = "/storage/:bucket/:key"
	storageBasePath   = "/storage/"
)

type mockStorageService struct {
	mock.Mock
}

func (m *mockStorageService) CreateBucket(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockStorageService) ListBuckets(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockStorageService) Upload(ctx context.Context, bucket, key string, content io.Reader) (*domain.Object, error) {
	args := m.Called(ctx, bucket, key, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Object), args.Error(1)
}

func (m *mockStorageService) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.Get(1).(*domain.Object), args.Error(2)
}

func (m *mockStorageService) ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error) {
	args := m.Called(ctx, bucket)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Object), args.Error(1)
}

func (m *mockStorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}

func (m *mockStorageService) GetObject(ctx context.Context, bucket, key string) (*domain.Object, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Object), args.Error(1)
}

func setupStorageHandlerTest(_ *testing.T) (*mockStorageService, *StorageHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	svc := new(mockStorageService)
	handler := NewStorageHandler(svc)
	r := gin.New()
	return svc, handler, r
}

func TestStorageHandlerUpload(t *testing.T) {
	svc, handler, r := setupStorageHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.PUT(storageObjectPath, handler.Upload)

	obj := &domain.Object{Key: testFileName, SizeBytes: 4}
	svc.On("Upload", mock.Anything, storageBucketName, testFileName, mock.Anything).Return(obj, nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", storageBasePath+storageBucketName+"/"+testFileName, strings.NewReader("data"))
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestStorageHandlerDownload(t *testing.T) {
	svc, handler, r := setupStorageHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET(storageObjectPath, handler.Download)

	content := io.NopCloser(bytes.NewBufferString("hello"))
	obj := &domain.Object{Key: testFileName, SizeBytes: 5, ContentType: "text/plain"}
	svc.On("Download", mock.Anything, storageBucketName, testFileName).Return(content, obj, nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", storageBasePath+storageBucketName+"/"+testFileName, nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello", w.Body.String())
}

func TestStorageHandlerList(t *testing.T) {
	svc, handler, r := setupStorageHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.GET("/storage/:bucket", handler.List)

	objects := []*domain.Object{{Key: testFileName, SizeBytes: 10}}
	svc.On("ListObjects", mock.Anything, storageBucketName).Return(objects, nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", storageBasePath+storageBucketName, nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStorageHandlerDelete(t *testing.T) {
	svc, handler, r := setupStorageHandlerTest(t)
	defer svc.AssertExpectations(t)

	r.DELETE(storageObjectPath, handler.Delete)

	svc.On("DeleteObject", mock.Anything, storageBucketName, testFileName).Return(nil)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", storageBasePath+storageBucketName+"/"+testFileName, nil)
	assert.NoError(t, err)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestStorageHandlerErrorPaths(t *testing.T) {
	setup := func(_ *testing.T) (*mockStorageService, *StorageHandler, *gin.Engine) {
		svc := new(mockStorageService)
		handler := NewStorageHandler(svc)
		r := gin.New()
		return svc, handler, r
	}

	t.Run("UploadMissingBucketOrKey", func(t *testing.T) {
		_, handler, r := setup(t)
		r.PUT(storageObjectPath, handler.Upload)
		req, _ := http.NewRequest("PUT", storageBasePath+storageBucketName, nil)
		// This URL matching might be tricky with params if one is empty.
		// Actually if I send /storage/b1 without key, it won't match :key path segment usually.
		// But handler logic says: bucket := c.Param("bucket"), key := c.Param("key").
		// If I use a route like /storage/:bucket/*key, maybe?
		// The existing route is /storage/:bucket/:key.
		// If I request /storage/b1/ then key is empty?
		// Let's try explicit empty key if possible or verify param logic.
		// Actually if Params are missing from context (e.g. if I don't use router but call handler directly), it fails.
		// But using router, if I request /storage/b1/, it might match if trailing slash enabled?
		// Let's rely on manually setting params for this specific test to ensure we hit the check.

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "bucket", Value: "b1"}, {Key: "key", Value: ""}}
		c.Request = req
		handler.Upload(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UploadServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		defer svc.AssertExpectations(t)
		r.PUT(storageObjectPath, handler.Upload)
		svc.On("Upload", mock.Anything, storageBucketName, testFileName, mock.Anything).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("PUT", storageBasePath+storageBucketName+"/"+testFileName, strings.NewReader("data"))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("DownloadServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		defer svc.AssertExpectations(t)
		r.GET(storageObjectPath, handler.Download)
		svc.On("Download", mock.Anything, storageBucketName, testFileName).Return(nil, nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", storageBasePath+storageBucketName+"/"+testFileName, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("ListServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		defer svc.AssertExpectations(t)
		r.GET("/storage/:bucket", handler.List)
		svc.On("ListObjects", mock.Anything, storageBucketName).Return(nil, errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("GET", storageBasePath+storageBucketName, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("DeleteServiceError", func(t *testing.T) {
		svc, handler, r := setup(t)
		defer svc.AssertExpectations(t)
		r.DELETE(storageObjectPath, handler.Delete)
		svc.On("DeleteObject", mock.Anything, storageBucketName, testFileName).Return(errors.New(errors.Internal, "error"))
		req, _ := http.NewRequest("DELETE", storageBasePath+storageBucketName+"/"+testFileName, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
