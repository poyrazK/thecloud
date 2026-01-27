package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	volumeAPIKey          = "test-key"
	volumeContentType     = "Content-Type"
	volumeApplicationJSON = "application/json"
	volumePath            = "/volumes"
	volumePathPrefix      = "/volumes/"
	volumeNewName         = "new-volume"
)

func TestClientListVolumes(t *testing.T) {
	mockVolumes := []Volume{
		{
			ID:     uuid.New(),
			Name:   "test-volume",
			SizeGB: 10,
			Status: "available",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, volumePath, r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set(volumeContentType, volumeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[[]Volume]{Data: mockVolumes})
	}))
	defer server.Close()

	client := NewClient(server.URL, volumeAPIKey)
	volumes, err := client.ListVolumes()

	assert.NoError(t, err)
	assert.Len(t, volumes, 1)
	assert.Equal(t, mockVolumes[0].ID, volumes[0].ID)
}

func TestClientCreateVolume(t *testing.T) {
	mockVolume := Volume{
		ID:     uuid.New(),
		Name:   volumeNewName,
		SizeGB: 20,
		Status: "creating",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, volumePath, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, volumeNewName, body["name"])
		assert.Equal(t, float64(20), body["size_gb"])

		w.Header().Set(volumeContentType, volumeApplicationJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response[Volume]{Data: mockVolume})
	}))
	defer server.Close()

	client := NewClient(server.URL, volumeAPIKey)
	volume, err := client.CreateVolume(volumeNewName, 20)

	assert.NoError(t, err)
	assert.Equal(t, mockVolume.ID, volume.ID)
	assert.Equal(t, volumeNewName, volume.Name)
}

func TestClientGetVolume(t *testing.T) {
	volID := uuid.New()
	mockVolume := Volume{
		ID:     volID,
		Name:   "test-volume",
		SizeGB: 10,
		Status: "available",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, volumePathPrefix+volID.String(), r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set(volumeContentType, volumeApplicationJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[Volume]{Data: mockVolume})
	}))
	defer server.Close()

	client := NewClient(server.URL, volumeAPIKey)
	volume, err := client.GetVolume(volID.String())

	assert.NoError(t, err)
	assert.Equal(t, volID, volume.ID)
}

func TestClientDeleteVolume(t *testing.T) {
	volID := uuid.New().String()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, volumePathPrefix+volID, r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, volumeAPIKey)
	err := client.DeleteVolume(volID)

	assert.NoError(t, err)
}

func TestClientVolumeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, volumeAPIKey)
	_, err := client.ListVolumes()
	assert.Error(t, err)

	_, err = client.CreateVolume("vol", 10)
	assert.Error(t, err)

	_, err = client.GetVolume(uuid.New().String())
	assert.Error(t, err)

	err = client.DeleteVolume(uuid.New().String())
	assert.Error(t, err)
}
