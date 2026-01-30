package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
)

const storageObjectPathFormat = "%s/storage/%s/%s"

func TestStorageE2E(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Storage E2E test: %v (server at %s not available)", err, testutil.TestBaseURL)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	token := registerAndLogin(t, client, "storage-test@thecloud.local", "Storage Tester")

	bucketName := fmt.Sprintf("bucket-%d", time.Now().UnixNano()%10000)
	key := "hello.txt"
	content := "Hello World E2E"

	// 1. Create Bucket
	t.Run("CreateBucket", func(t *testing.T) {
		reqBody := map[string]interface{}{"name": bucketName, "is_public": false}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", testutil.TestBaseURL+"/storage/buckets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 2. Upload Object
	t.Run("UploadObject", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", fmt.Sprintf(storageObjectPathFormat, testutil.TestBaseURL, bucketName, key), bytes.NewBufferString(content))
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 3. Download Object
	t.Run("DownloadObject", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf(storageObjectPathFormat, testutil.TestBaseURL, bucketName, key), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	// 4. Enable Versioning
	t.Run("EnableVersioning", func(t *testing.T) {
		reqBody := map[string]bool{"enabled": true}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/storage/buckets/%s/versioning", testutil.TestBaseURL, bucketName), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", testutil.TestContentTypeAppJSON)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 5. Upload second version
	var version2ID string
	t.Run("UploadSecondVersion", func(t *testing.T) {
		newContent := "Hello World Version 2"
		req, _ := http.NewRequest("PUT", fmt.Sprintf(storageObjectPathFormat, testutil.TestBaseURL, bucketName, key), bytes.NewBufferString(newContent))
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// The response wrapper might be different, let's check httputil/Success
		// Usually Success(c, code, data) wraps it in {"data": data}
		type SuccessResp struct {
			Data domain.Object `json:"data"`
		}
		var s SuccessResp
		err = json.NewDecoder(resp.Body).Decode(&s)
		require.NoError(t, err)
		version2ID = s.Data.VersionID
		assert.NotEmpty(t, version2ID)
	})

	// 6. List Versions
	t.Run("ListVersions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/storage/versions/%s/%s", testutil.TestBaseURL, bucketName, key), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		type ListResp struct {
			Data []domain.Object `json:"data"`
		}
		var l ListResp
		err = json.NewDecoder(resp.Body).Decode(&l)
		require.NoError(t, err)
		assert.True(t, len(l.Data) >= 2)
	})

	// 7. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		// Delete object
		req, _ := http.NewRequest("DELETE", fmt.Sprintf(storageObjectPathFormat, testutil.TestBaseURL, bucketName, key), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)
		resp, _ := client.Do(req)
		_ = resp.Body.Close()

		// Delete bucket
		req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/storage/buckets/%s", testutil.TestBaseURL, bucketName), nil)
		req.Header.Set(testutil.TestHeaderAPIKey, token)
		applyTenantHeader(t, req, token)
		resp, _ = client.Do(req)
		_ = resp.Body.Close()
	})
}
