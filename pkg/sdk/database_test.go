package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	dbTestName        = "test-db"
	dbID              = "db-123"
	dbAPIKey           = "test-api-key"
	dbContentType      = "Content-Type"
	dbApplicationJSON  = "application/json"
	dbPathPrefix       = "/databases/"
)

func TestClientCreateDatabase(t *testing.T) {
	vpcID := "vpc-123"
	expectedDB := Database{
		ID:      "db-1",
		Name:    dbTestName,
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

		w.Header().Set(dbContentType, dbApplicationJSON)
		resp := Response[Database]{Data: expectedDB}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, dbAPIKey)
	db, err := client.CreateDatabase(dbTestName, "postgres", "14", &vpcID)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, expectedDB.ID, db.ID)
}

func TestClientListDatabases(t *testing.T) {
	expectedDBs := []*Database{
		{ID: "db-1", Name: "db-1"},
		{ID: "db-2", Name: "db-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/databases", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(dbContentType, dbApplicationJSON)
		resp := Response[[]*Database]{Data: expectedDBs}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, dbAPIKey)
	dbs, err := client.ListDatabases()

	assert.NoError(t, err)
	assert.Len(t, dbs, 2)
}


func TestClientGetDatabase(t *testing.T) {
	id := dbID
	expectedDB := Database{
		ID:   id,
		Name: dbTestName,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, dbPathPrefix+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(dbContentType, dbApplicationJSON)
		resp := Response[Database]{Data: expectedDB}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, dbAPIKey)
	db, err := client.GetDatabase(id)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, expectedDB.ID, db.ID)
}


func TestClientDeleteDatabase(t *testing.T) {
	id := dbID

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, dbPathPrefix+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, dbAPIKey)
	err := client.DeleteDatabase(id)

	assert.NoError(t, err)
}


func TestClientGetDatabaseConnectionString(t *testing.T) {
	id := dbID
	connStr := "postgres://user:pass@host:5432/dbname"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, dbPathPrefix+id+"/connection", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(dbContentType, dbApplicationJSON)
		resp := Response[map[string]string]{Data: map[string]string{"connection_string": connStr}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, dbAPIKey)
	result, err := client.GetDatabaseConnectionString(id)

	assert.NoError(t, err)
	assert.Equal(t, connStr, result)
}

func TestClientDatabaseErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, dbAPIKey)

	_, err := client.CreateDatabase("db", "postgres", "14", nil)
	assert.Error(t, err)

	_, err = client.ListDatabases()
	assert.Error(t, err)

	_, err = client.GetDatabase("db-1")
	assert.Error(t, err)

	err = client.DeleteDatabase("db-1")
	assert.Error(t, err)

	_, err = client.GetDatabaseConnectionString("db-1")
	assert.Error(t, err)
}
