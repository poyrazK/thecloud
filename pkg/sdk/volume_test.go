package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestClient_ListVolumes(t *testing.T) {
	mockVolumes := []Volume{
		{
			ID:     uuid.New(),
			Name:   "test-volume",
			SizeGB: 10,
			Status: "available",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/volumes", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[[]Volume]{Data: mockVolumes})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	volumes, err := client.ListVolumes()

	assert.NoError(t, err)
	assert.Len(t, volumes, 1)
	assert.Equal(t, mockVolumes[0].ID, volumes[0].ID)
}

func TestClient_CreateVolume(t *testing.T) {
	mockVolume := Volume{
		ID:     uuid.New(),
		Name:   "new-volume",
		SizeGB: 20,
		Status: "creating",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/volumes", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "new-volume", body["name"])
		assert.Equal(t, float64(20), body["size_gb"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response[Volume]{Data: mockVolume})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	volume, err := client.CreateVolume("new-volume", 20)

	assert.NoError(t, err)
	assert.Equal(t, mockVolume.ID, volume.ID)
	assert.Equal(t, "new-volume", volume.Name)
}

func TestClient_GetVolume(t *testing.T) {
	volID := uuid.New()
	mockVolume := Volume{
		ID:     volID,
		Name:   "test-volume",
		SizeGB: 10,
		Status: "available",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/volumes/"+volID.String(), r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[Volume]{Data: mockVolume})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	volume, err := client.GetVolume(volID.String())

	assert.NoError(t, err)
	assert.Equal(t, volID, volume.ID)
}

func TestClient_DeleteVolume(t *testing.T) {
	volID := uuid.New().String()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/volumes/"+volID, r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	err := client.DeleteVolume(volID)

	assert.NoError(t, err)
}

func TestClient_VolumeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	_, err := client.ListVolumes()
	assert.Error(t, err)

	_, err = client.CreateVolume("vol", 10)
	assert.Error(t, err)

	_, err = client.GetVolume(uuid.New().String())
	assert.Error(t, err)

	err = client.DeleteVolume(uuid.New().String())
	assert.Error(t, err)
}
