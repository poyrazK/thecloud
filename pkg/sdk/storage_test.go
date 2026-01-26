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
	obj, err := client.UploadObject(bucket, key, strings.NewReader(content))

	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func TestClient_UploadObjectErrorStatus(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	_, err := client.UploadObject(bucket, key, strings.NewReader("data"))

	assert.Error(t, err)
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

func TestClient_ListVersions(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"
	expected := []Object{{ID: "obj-1", VersionID: "v1"}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/versions/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]Object]{Data: expected})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	versions, err := client.ListVersions(bucket, key)

	assert.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "v1", versions[0].VersionID)
}

func TestClient_DownloadObjectWithVersion(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"
	version := "v1"
	content := "hello world"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, version, r.URL.Query().Get("versionId"))
		w.Write([]byte(content))
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	readCloser, err := client.DownloadObject(bucket, key, version)

	assert.NoError(t, err)
	defer readCloser.Close()

	body, err := io.ReadAll(readCloser)
	assert.NoError(t, err)
	assert.Equal(t, content, string(body))
}

func TestClient_DownloadObjectErrorStatus(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket+"/"+key, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	_, err := client.DownloadObject(bucket, key)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestClient_DeleteObjectWithVersion(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"
	version := "v1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, version, r.URL.Query().Get("versionId"))
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.DeleteObject(bucket, key, version)

	assert.NoError(t, err)
}

func TestClient_CreateBucket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/buckets", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var payload struct {
			Name     string `json:"name"`
			IsPublic bool   `json:"is_public"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, "bucket-1", payload.Name)
		assert.True(t, payload.IsPublic)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[Bucket]{Data: Bucket{ID: "b1", Name: "bucket-1", IsPublic: true}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	bucket, err := client.CreateBucket("bucket-1", true)

	assert.NoError(t, err)
	assert.Equal(t, "b1", bucket.ID)
}

func TestClient_ListBuckets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/buckets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]Bucket]{Data: []Bucket{{ID: "b1", Name: "bucket-1"}}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	buckets, err := client.ListBuckets()

	assert.NoError(t, err)
	assert.Len(t, buckets, 1)
	assert.Equal(t, "bucket-1", buckets[0].Name)
}

func TestClient_DeleteBucket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/buckets/bucket-1", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.DeleteBucket("bucket-1")

	assert.NoError(t, err)
}

func TestClient_SetBucketVersioning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/buckets/bucket-1/versioning", r.URL.Path)
		assert.Equal(t, http.MethodPatch, r.Method)

		var payload struct {
			Enabled bool `json:"enabled"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.True(t, payload.Enabled)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.SetBucketVersioning("bucket-1", true)

	assert.NoError(t, err)
}

func TestClient_GetStorageClusterStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/cluster/status", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[StorageCluster]{Data: StorageCluster{Nodes: []StorageNode{{ID: "node-1"}}}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	status, err := client.GetStorageClusterStatus()

	assert.NoError(t, err)
	assert.Len(t, status.Nodes, 1)
	assert.Equal(t, "node-1", status.Nodes[0].ID)
}

func TestClient_GeneratePresignedURL(t *testing.T) {
	bucket := "my-bucket"
	key := "file.txt"
	method := http.MethodGet

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/storage/presign/"+bucket+"/"+key, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var payload struct {
			Method    string `json:"method"`
			ExpirySec int    `json:"expiry_seconds"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, method, payload.Method)
		assert.Equal(t, 60, payload.ExpirySec)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[PresignedURL]{Data: PresignedURL{URL: "http://example.com", Method: method}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	url, err := client.GeneratePresignedURL(bucket, key, method, 60)

	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", url.URL)
}

func TestClient_LifecycleRules(t *testing.T) {
	bucket := "my-bucket"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/buckets/"+bucket+"/lifecycle":
			var payload struct {
				Prefix         string `json:"prefix"`
				ExpirationDays int    `json:"expiration_days"`
				Enabled        bool   `json:"enabled"`
			}
			err := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.Equal(t, "logs/", payload.Prefix)
			assert.Equal(t, 30, payload.ExpirationDays)
			assert.True(t, payload.Enabled)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Response[LifecycleRule]{Data: LifecycleRule{ID: "rule-1", BucketName: bucket}})
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/"+bucket+"/lifecycle":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Response[[]LifecycleRule]{Data: []LifecycleRule{{ID: "rule-1", BucketName: bucket}}})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/"+bucket+"/lifecycle/rule-1":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	created, err := client.CreateLifecycleRule(bucket, "logs/", 30, true)
	assert.NoError(t, err)
	assert.Equal(t, "rule-1", created.ID)

	rules, err := client.ListLifecycleRules(bucket)
	assert.NoError(t, err)
	assert.Len(t, rules, 1)

	err = client.DeleteLifecycleRule(bucket, "rule-1")
	assert.NoError(t, err)
}

func TestClient_StorageListErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	_, err := client.ListObjects("bucket")
	assert.Error(t, err)

	_, err = client.ListVersions("bucket", "key")
	assert.Error(t, err)

	_, err = client.ListBuckets()
	assert.Error(t, err)

	_, err = client.GetStorageClusterStatus()
	assert.Error(t, err)

	_, err = client.ListLifecycleRules("bucket")
	assert.Error(t, err)
}
