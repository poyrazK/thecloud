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
