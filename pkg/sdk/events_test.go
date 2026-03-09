package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientListEvents(t *testing.T) {
	expectedEvents := []Event{
		{
			ID:           uuid.New(),
			Action:       "instance.created",
			ResourceID:   "inst-123",
			ResourceType: "instance",
			Metadata:     json.RawMessage(`{"name":"test"}`),
			CreatedAt:    time.Now(),
		},
		{
			ID:           uuid.New(),
			Action:       "instance.terminated",
			ResourceID:   "inst-456",
			ResourceType: "instance",
			Metadata:     json.RawMessage(`{}`),
			CreatedAt:    time.Now(),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/events", r.URL.Path)
		assert.Equal(t, "limit=50", r.URL.RawQuery)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[[]Event]{Data: expectedEvents}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	events, err := client.ListEvents()

	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, "instance.created", events[0].Action)
}

func TestClientListEventsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	_, err := client.ListEvents()

	require.Error(t, err)
}
