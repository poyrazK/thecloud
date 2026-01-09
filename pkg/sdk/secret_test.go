package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_CreateSecret(t *testing.T) {
	expectedSecret := Secret{
		ID:             "sec-1",
		UserID:         "user-1",
		Name:           "test-secret",
		EncryptedValue: "test-value",
		Description:    "test description",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req CreateSecretInput
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, expectedSecret.Name, req.Name)
		assert.Equal(t, "test-value", req.Value)
		assert.Equal(t, expectedSecret.Description, req.Description)

		w.Header().Set("Content-Type", "application/json")
		// Wrap in Response[Secret]
		resp := Response[Secret]{Data: expectedSecret}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	secret, err := client.CreateSecret("test-secret", "test-value", "test description")

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, expectedSecret.ID, secret.ID)
	assert.Equal(t, expectedSecret.EncryptedValue, secret.EncryptedValue)
}

func TestClient_ListSecrets(t *testing.T) {
	expectedSecrets := []*Secret{
		{ID: "sec-1", Name: "secret-1"},
		{ID: "sec-2", Name: "secret-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[[]*Secret]{Data: expectedSecrets}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	secrets, err := client.ListSecrets()

	assert.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.Equal(t, expectedSecrets[0].Name, secrets[0].Name)
}

func TestClient_GetSecret(t *testing.T) {
	id := "sec-1"
	expectedSecret := Secret{ID: id, Name: "test-secret"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[Secret]{Data: expectedSecret}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	secret, err := client.GetSecret(id)

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, expectedSecret.ID, secret.ID)
}

func TestClient_DeleteSecret(t *testing.T) {
	id := "sec-1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteSecret(id)

	assert.NoError(t, err)
}
