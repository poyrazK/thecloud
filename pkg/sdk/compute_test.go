package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_ListInstances(t *testing.T) {
	mockInstances := []Instance{
		{
			ID:     "inst-1",
			Name:   "test-instance",
			Status: "running",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[[]Instance]{Data: mockInstances})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	instances, err := client.ListInstances()

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "inst-1", instances[0].ID)
}

func TestClient_GetInstance(t *testing.T) {
	mockInstance := Instance{
		ID:     "inst-1",
		Name:   "test-instance",
		Status: "running",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances/inst-1", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[Instance]{Data: mockInstance})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	instance, err := client.GetInstance("inst-1")

	assert.NoError(t, err)
	assert.Equal(t, "inst-1", instance.ID)
}

func TestClient_LaunchInstance(t *testing.T) {
	mockInstance := Instance{
		ID:     "inst-1",
		Name:   "new-instance",
		Status: "running",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "new-instance", body["name"])
		assert.Equal(t, "nginx", body["image"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response[Instance]{Data: mockInstance})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	instance, err := client.LaunchInstance("new-instance", "nginx", "80:80", "", "", nil)

	assert.NoError(t, err)
	assert.Equal(t, "inst-1", instance.ID)
	assert.Equal(t, "new-instance", instance.Name)
}

func TestClient_StopInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances/inst-1/stop", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "application/json") // Added for consistency, though no body is returned
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	err := client.StopInstance("inst-1")

	assert.NoError(t, err)
}

func TestClient_TerminateInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances/inst-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	err := client.TerminateInstance("inst-1")

	assert.NoError(t, err)
}

func TestClient_GetInstanceLogs(t *testing.T) {
	mockLogs := "hello world\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances/inst-1/logs", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockLogs))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	logs, err := client.GetInstanceLogs("inst-1")

	assert.NoError(t, err)
	assert.Equal(t, mockLogs, logs)
}

func TestClient_GetInstanceLogsErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances/inst-1/logs", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	_, err := client.GetInstanceLogs("inst-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
}

func TestClient_GetInstanceLogsRequestError(t *testing.T) {
	client := NewClient("http://127.0.0.1:0", "test-key")
	_, err := client.GetInstanceLogs("inst-1")

	assert.Error(t, err)
}

func TestClient_GetInstanceStats(t *testing.T) {
	mockStats := InstanceStats{
		CPUPercentage:    15.5,
		MemoryUsageBytes: 1024 * 1024 * 10,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/instances/inst-1/stats", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[InstanceStats]{Data: mockStats})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	stats, err := client.GetInstanceStats("inst-1")

	assert.NoError(t, err)
	assert.Equal(t, 15.5, stats.CPUPercentage)
}

func TestClient_ComputeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	_, err := client.GetInstance("inst-1")
	assert.Error(t, err)

	_, err = client.GetInstanceStats("inst-1")
	assert.Error(t, err)
}

func TestClient_LaunchInstanceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	_, err := client.LaunchInstance("name", "img", "80", "", "", nil)
	assert.Error(t, err)
}

func TestClient_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"type": "bad_request", "message": "invalid input"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	_, err := client.ListInstances()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api error")
	assert.Contains(t, err.Error(), "invalid input")
}
