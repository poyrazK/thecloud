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

const (
	iacTestStackName     = "test-stack"
	iacTestTemplate      = "version: 1.0\nresources: []"
	iacTestAPIKey        = "test-api-key"
	iacTestContentType   = "Content-Type"
	iacTestAppJSON       = "application/json"
	iacTestStackOneName  = "stack-1"
	iacTestStackTwoName  = "stack-2"
)

func TestClientCreateStack(t *testing.T) {
	expectedStack := domain.Stack{
		ID:        uuid.New(),
		Name:      iacTestStackName,
		Template:  iacTestTemplate,
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

		w.Header().Set(iacTestContentType, iacTestAppJSON)
		json.NewEncoder(w).Encode(expectedStack)
	}))
	defer server.Close()

	client := NewClient(server.URL, iacTestAPIKey)
	stack, err := client.CreateStack(iacTestStackName, iacTestTemplate, map[string]string{})

	assert.NoError(t, err)
	assert.NotNil(t, stack)
	assert.Equal(t, expectedStack.ID, stack.ID)
	assert.Equal(t, expectedStack.Name, stack.Name)
}

func TestClientListStacks(t *testing.T) {
	expectedStacks := []*domain.Stack{
		{ID: uuid.New(), Name: iacTestStackOneName},
		{ID: uuid.New(), Name: iacTestStackTwoName},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/stacks", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(iacTestContentType, iacTestAppJSON)
		json.NewEncoder(w).Encode(expectedStacks)
	}))
	defer server.Close()

	client := NewClient(server.URL, iacTestAPIKey)
	stacks, err := client.ListStacks()

	assert.NoError(t, err)
	assert.Len(t, stacks, 2)
	assert.Equal(t, expectedStacks[0].Name, stacks[0].Name)
}

func TestClientGetStack(t *testing.T) {
	id := uuid.New().String()
	expectedStack := domain.Stack{
		ID:   uuid.MustParse(id),
		Name: iacTestStackName,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/stacks/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(iacTestContentType, iacTestAppJSON)
		json.NewEncoder(w).Encode(expectedStack)
	}))
	defer server.Close()

	client := NewClient(server.URL, iacTestAPIKey)
	stack, err := client.GetStack(id)

	assert.NoError(t, err)
	assert.NotNil(t, stack)
	assert.Equal(t, expectedStack.ID, stack.ID)
}

func TestClientDeleteStack(t *testing.T) {
	id := uuid.New().String()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/iac/stacks/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, iacTestAPIKey)
	err := client.DeleteStack(id)

	assert.NoError(t, err)
}

func TestClientValidateTemplate(t *testing.T) {
	template := iacTestTemplate
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

		w.Header().Set(iacTestContentType, iacTestAppJSON)
		json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL, iacTestAPIKey)
	resp, err := client.ValidateTemplate(template)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Valid)
}

func TestClientIacErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, iacTestAPIKey)
	_, err := client.CreateStack("stack", "template", map[string]string{})
	assert.Error(t, err)

	_, err = client.ListStacks()
	assert.Error(t, err)

	_, err = client.GetStack(iacTestStackOneName)
	assert.Error(t, err)

	err = client.DeleteStack(iacTestStackOneName)
	assert.Error(t, err)

	_, err = client.ValidateTemplate("template")
	assert.Error(t, err)
}
