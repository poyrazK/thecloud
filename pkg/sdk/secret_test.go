package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	secretTestID          = "sec-1"
	secretTestUserID      = "user-1"
	secretTestName        = "test-secret"
	secretTestValue       = "test-value"
	secretTestDescription = "test description"
	secretTestAPIKey      = "test-api-key"
	secretTestContentType = "Content-Type"
	secretTestAppJSON     = "application/json"
)

func TestClientCreateSecret(t *testing.T) {
	expectedSecret := Secret{
		ID:             secretTestID,
		UserID:         secretTestUserID,
		Name:           secretTestName,
		EncryptedValue: secretTestValue,
		Description:    secretTestDescription,
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
		assert.Equal(t, secretTestValue, req.Value)
		assert.Equal(t, expectedSecret.Description, req.Description)

		w.Header().Set(secretTestContentType, secretTestAppJSON)
		// Wrap in Response[Secret]
		resp := Response[Secret]{Data: expectedSecret}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, secretTestAPIKey)
	secret, err := client.CreateSecret(secretTestName, secretTestValue, secretTestDescription)

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, expectedSecret.ID, secret.ID)
	assert.Equal(t, expectedSecret.EncryptedValue, secret.EncryptedValue)
}

func TestClientListSecrets(t *testing.T) {
	expectedSecrets := []*Secret{
		{ID: "sec-1", Name: "secret-1"},
		{ID: "sec-2", Name: "secret-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(secretTestContentType, secretTestAppJSON)
		resp := Response[[]*Secret]{Data: expectedSecrets}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, secretTestAPIKey)
	secrets, err := client.ListSecrets()

	assert.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.Equal(t, expectedSecrets[0].Name, secrets[0].Name)
}

func TestClientGetSecret(t *testing.T) {
	expectedSecret := Secret{ID: secretTestID, Name: secretTestName}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets/"+secretTestID, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(secretTestContentType, secretTestAppJSON)
		resp := Response[Secret]{Data: expectedSecret}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, secretTestAPIKey)
	secret, err := client.GetSecret(secretTestID)

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, expectedSecret.ID, secret.ID)
}

func TestClientDeleteSecret(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secrets/"+secretTestID, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, secretTestAPIKey)
	err := client.DeleteSecret(secretTestID)

	assert.NoError(t, err)
}

func TestClientSecretErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, secretTestAPIKey)
	_, err := client.CreateSecret("secret", "value", "desc")
	assert.Error(t, err)

	_, err = client.ListSecrets()
	assert.Error(t, err)

	_, err = client.GetSecret(secretTestID)
	assert.Error(t, err)

	err = client.DeleteSecret(secretTestID)
	assert.Error(t, err)
}
