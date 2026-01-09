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

func TestClient_CreateStack(t *testing.T) {
	expectedStack := domain.Stack{
		ID:        uuid.New(),
		Name:      "test-stack",
		Template:  "version: 1.0\nresources: []",
		Status:    domain.StackStatusCreateInProgress,
		CreatedAt: time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/stacks", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, expectedStack.Name, req["name"])
		assert.Equal(t, expectedStack.Template, req["template"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedStack)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	stack, err := client.CreateStack("test-stack", "version: 1.0\nresources: []", map[string]string{})

	assert.NoError(t, err)
	assert.NotNil(t, stack)
	assert.Equal(t, expectedStack.ID, stack.ID)
	assert.Equal(t, expectedStack.Name, stack.Name)
}

func TestClient_ListStacks(t *testing.T) {
	expectedStacks := []*domain.Stack{
		{ID: uuid.New(), Name: "stack-1"},
		{ID: uuid.New(), Name: "stack-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/stacks", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedStacks)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	stacks, err := client.ListStacks()

	assert.NoError(t, err)
	assert.Len(t, stacks, 2)
	assert.Equal(t, expectedStacks[0].Name, stacks[0].Name)
}

func TestClient_GetStack(t *testing.T) {
	id := uuid.New().String()
	expectedStack := domain.Stack{
		ID:   uuid.MustParse(id),
		Name: "test-stack",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/stacks/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedStack)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	stack, err := client.GetStack(id)

	assert.NoError(t, err)
	assert.NotNil(t, stack)
	assert.Equal(t, expectedStack.ID, stack.ID)
}

func TestClient_DeleteStack(t *testing.T) {
	id := uuid.New().String()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/stacks/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteStack(id)

	assert.NoError(t, err)
}

func TestClient_ValidateTemplate(t *testing.T) {
	template := "version: 1.0\nresources: []"
	expectedResp := domain.TemplateValidateResponse{
		Valid:  true,
		Errors: []string{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/validate", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, template, req["template"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	resp, err := client.ValidateTemplate(template)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Valid)
}
