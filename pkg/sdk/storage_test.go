package sdk

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_ListObjects(t *testing.T) {
	bucket := "my-bucket"
	expectedObjects := []Object{
		{ID: "obj-1", Key: "file1.txt", Bucket: bucket, CreatedAt: time.Now()},
		{ID: "obj-2", Key: "file2.txt", Bucket: bucket, CreatedAt: time.Now()},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[[]Object]{Data: expectedObjects}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	objects, err := client.ListObjects(bucket)

	assert.NoError(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, expectedObjects[0].Key, objects[0].Key)
}

func TestClient_UploadObject(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"
	content := "hello world"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, content, string(body))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.UploadObject(bucket, key, strings.NewReader(content))

	assert.NoError(t, err)
}

func TestClient_DownloadObject(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"
	content := "hello world"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Write([]byte(content))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	readCloser, err := client.DownloadObject(bucket, key)

	assert.NoError(t, err)
	defer readCloser.Close()

	body, err := io.ReadAll(readCloser)
	assert.NoError(t, err)
	assert.Equal(t, content, string(body))
}

func TestClient_DeleteObject(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteObject(bucket, key)

	assert.NoError(t, err)
}
