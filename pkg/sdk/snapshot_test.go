package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestClient_CreateSnapshot(t *testing.T) {
	volumeID := uuid.New()
	expectedSnapshot := domain.Snapshot{
		ID:          uuid.New(),
		VolumeID:    volumeID,
		Description: "test snapshot",
		Status:      domain.SnapshotStatusAvailable,
		CreatedAt:   time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/snapshots", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, volumeID.String(), req["volume_id"])
		assert.Equal(t, expectedSnapshot.Description, req["description"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedSnapshot)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	snapshot, err := client.CreateSnapshot(volumeID, "test snapshot")

	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, expectedSnapshot.ID, snapshot.ID)
	assert.Equal(t, expectedSnapshot.VolumeID, snapshot.VolumeID)
}

func TestClient_ListSnapshots(t *testing.T) {
	expectedSnapshots := []*domain.Snapshot{
		{ID: uuid.New(), Description: "snap-1"},
		{ID: uuid.New(), Description: "snap-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/snapshots", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedSnapshots)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	snapshots, err := client.ListSnapshots()

	assert.NoError(t, err)
	assert.Len(t, snapshots, 2)
	assert.Equal(t, expectedSnapshots[0].Description, snapshots[0].Description)
}

func TestClient_GetSnapshot(t *testing.T) {
	id := uuid.New().String()
	expectedSnapshot := domain.Snapshot{ID: uuid.MustParse(id), Description: "test snapshot"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/snapshots/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedSnapshot)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	snapshot, err := client.GetSnapshot(id)

	assert.NoError(t, err)
	assert.NotNil(t, snapshot)
	assert.Equal(t, expectedSnapshot.ID, snapshot.ID)
}

func TestClient_DeleteSnapshot(t *testing.T) {
	id := uuid.New().String()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/snapshots/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteSnapshot(id)

	assert.NoError(t, err)
}

func TestClient_RestoreSnapshot(t *testing.T) {
	id := uuid.New().String()
	newVolumeName := "restored-volume"
	expectedVolume := domain.Volume{
		ID:   uuid.New(),
		Name: newVolumeName,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/snapshots/"+id+"/restore", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, newVolumeName, req["new_volume_name"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedVolume)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	volume, err := client.RestoreSnapshot(id, newVolumeName)

	assert.NoError(t, err)
	assert.NotNil(t, volume)
	assert.Equal(t, expectedVolume.ID, volume.ID)
	assert.Equal(t, expectedVolume.Name, volume.Name)
}

func TestClient_SnapshotErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	_, err := client.CreateSnapshot(uuid.New(), "snap")
	assert.Error(t, err)

	_, err = client.ListSnapshots()
	assert.Error(t, err)

	_, err = client.GetSnapshot("snap-1")
	assert.Error(t, err)

	err = client.DeleteSnapshot("snap-1")
	assert.Error(t, err)

	_, err = client.RestoreSnapshot("snap-1", "vol")
	assert.Error(t, err)
}
