package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClientListVPCs(t *testing.T) {
	mockVPCs := []VPC{
		{
			ID:   "vpc-1",
			Name: "test-vpc",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[[]VPC]{Data: mockVPCs})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	vpcs, err := client.ListVPCs()

	assert.NoError(t, err)
	assert.Len(t, vpcs, 1)
	assert.Equal(t, "vpc-1", vpcs[0].ID)
}

func TestClientCreateVPC(t *testing.T) {
	mockVPC := VPC{
		ID:   "vpc-1",
		Name: testVpcName,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, testVpcName, body["name"])
		assert.Equal(t, testutil.TestCIDR, body["cidr_block"])

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Response[VPC]{Data: mockVPC})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	vpc, err := client.CreateVPC(testVpcName, testutil.TestCIDR)

	assert.NoError(t, err)
	assert.Equal(t, "vpc-1", vpc.ID)
	assert.Equal(t, testVpcName, vpc.Name)
}

func TestClientGetVPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs/vpc-1", r.URL.Path)
		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[VPC]{Data: VPC{ID: "vpc-1"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	vpc, err := client.GetVPC("vpc-1")

	assert.NoError(t, err)
	assert.Equal(t, "vpc-1", vpc.ID)
}

func TestClientDeleteVPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs/vpc-1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.DeleteVPC("vpc-1")

	assert.NoError(t, err)
}
