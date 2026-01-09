package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_ListSubnets(t *testing.T) {
	vpcID := "vpc-123"
	expectedSubnets := []*Subnet{
		{ID: "subnet-1", Name: "subnet-1", VpcID: vpcID, CIDRBlock: "10.0.1.0/24"},
		{ID: "subnet-2", Name: "subnet-2", VpcID: vpcID, CIDRBlock: "10.0.2.0/24"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs/"+vpcID+"/subnets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[[]*Subnet]{Data: expectedSubnets}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	subnets, err := client.ListSubnets(vpcID)

	assert.NoError(t, err)
	assert.Len(t, subnets, 2)
	assert.Equal(t, expectedSubnets[0].CIDRBlock, subnets[0].CIDRBlock)
}

func TestClient_CreateSubnet(t *testing.T) {
	vpcID := "vpc-123"
	expectedSubnet := &Subnet{
		ID:        "subnet-1",
		VpcID:     vpcID,
		Name:      "test-subnet",
		CIDRBlock: "10.0.1.0/24",
		AZ:        "us-east-1a",
		Status:    "available",
		CreatedAt: time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs/"+vpcID+"/subnets", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, expectedSubnet.Name, req["name"])
		assert.Equal(t, expectedSubnet.CIDRBlock, req["cidr_block"])
		assert.Equal(t, expectedSubnet.AZ, req["availability_zone"])

		w.Header().Set("Content-Type", "application/json")
		resp := Response[*Subnet]{Data: expectedSubnet}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	subnet, err := client.CreateSubnet(vpcID, "test-subnet", "10.0.1.0/24", "us-east-1a")

	assert.NoError(t, err)
	assert.NotNil(t, subnet)
	assert.Equal(t, expectedSubnet.ID, subnet.ID)
	assert.Equal(t, expectedSubnet.CIDRBlock, subnet.CIDRBlock)
}

func TestClient_DeleteSubnet(t *testing.T) {
	id := "subnet-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subnets/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	err := client.DeleteSubnet(id)

	assert.NoError(t, err)
}

func TestClient_GetSubnet(t *testing.T) {
	id := "subnet-123"
	expectedSubnet := &Subnet{
		ID:        id,
		Name:      "test-subnet",
		CIDRBlock: "10.0.1.0/24",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subnets/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		resp := Response[*Subnet]{Data: expectedSubnet}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key")
	subnet, err := client.GetSubnet(id)

	assert.NoError(t, err)
	assert.NotNil(t, subnet)
	assert.Equal(t, expectedSubnet.ID, subnet.ID)
	assert.Equal(t, expectedSubnet.CIDRBlock, subnet.CIDRBlock)
}
