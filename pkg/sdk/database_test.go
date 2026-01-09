package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_CreateDatabase(t *testing.T) {
	vpcID := "vpc-123"
	expectedDB := Database{
		ID:      "db-1",
		Name:    "test-db",
		Engine:  "postgres",
		Version: "14",
		Status:  "CREATING",
		VpcID:   &vpcID,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/databases", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req CreateDatabaseInput
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, expectedDB.Name, req.Name)
		assert.Equal(t, expectedDB.Engine, req.Engine)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[Database]{Data: expectedDB}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	db, err := client.CreateDatabase("test-db", "postgres", "14", &vpcID)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, expectedDB.ID, db.ID)
}

func TestClient_ListDatabases(t *testing.T) {
	expectedDBs := []*Database{
		{ID: "db-1", Name: "db-1"},
		{ID: "db-2", Name: "db-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/databases", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[[]*Database]{Data: expectedDBs}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	dbs, err := client.ListDatabases()

	assert.NoError(t, err)
	assert.Len(t, dbs, 2)
}

func TestClient_GetDatabase(t *testing.T) {
	id := "db-123"
	expectedDB := Database{
		ID:   id,
		Name: "test-db",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/databases/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[Database]{Data: expectedDB}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	db, err := client.GetDatabase(id)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, expectedDB.ID, db.ID)
}

func TestClient_DeleteDatabase(t *testing.T) {
	id := "db-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/databases/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteDatabase(id)

	assert.NoError(t, err)
}

func TestClient_GetDatabaseConnectionString(t *testing.T) {
	id := "db-123"
	connStr := "postgres://user:pass@host:5432/dbname"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/databases/"+id+"/connection", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[map[string]string]{Data: map[string]string{"connection_string": connStr}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	result, err := client.GetDatabaseConnectionString(id)

	assert.NoError(t, err)
	assert.Equal(t, connStr, result)
}
