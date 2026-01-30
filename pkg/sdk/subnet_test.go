package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClientListSubnets(t *testing.T) {
	vpcID := "vpc-123"
	expectedSubnets := []*Subnet{
		{ID: subnet1, Name: subnet1, VpcID: vpcID, CIDRBlock: testutil.TestSubnetCIDR},
		{ID: "subnet-2", Name: "subnet-2", VpcID: vpcID, CIDRBlock: testutil.TestSubnet2CIDR},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/vpcs/"+vpcID+"/subnets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		resp := Response[[]*Subnet]{Data: expectedSubnets}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	subnets, err := client.ListSubnets(vpcID)

	assert.NoError(t, err)
	assert.Len(t, subnets, 2)
	assert.Equal(t, expectedSubnets[0].CIDRBlock, subnets[0].CIDRBlock)
}

func TestClientCreateSubnet(t *testing.T) {
	vpcID := "vpc-123"
	expectedSubnet := &Subnet{
		ID:        subnet1,
		VpcID:     vpcID,
		Name:      testSubnetName,
		CIDRBlock: testutil.TestSubnetCIDR,
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

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		resp := Response[*Subnet]{Data: expectedSubnet}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	subnet, err := client.CreateSubnet(vpcID, testSubnetName, testutil.TestSubnetCIDR, "us-east-1a")

	assert.NoError(t, err)
	assert.NotNil(t, subnet)
	assert.Equal(t, expectedSubnet.ID, subnet.ID)
	assert.Equal(t, expectedSubnet.CIDRBlock, subnet.CIDRBlock)
}

func TestClientDeleteSubnet(t *testing.T) {
	id := "subnet-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subnets/"+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.DeleteSubnet(id)

	assert.NoError(t, err)
}

func TestClientGetSubnet(t *testing.T) {
	id := "subnet-123"
	expectedSubnet := &Subnet{
		ID:        id,
		Name:      testSubnetName,
		CIDRBlock: testutil.TestSubnetCIDR,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subnets/"+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", testutil.TestContentTypeAppJSON)
		resp := Response[*Subnet]{Data: expectedSubnet}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	subnet, err := client.GetSubnet(id)

	assert.NoError(t, err)
	assert.NotNil(t, subnet)
	assert.Equal(t, expectedSubnet.ID, subnet.ID)
	assert.Equal(t, expectedSubnet.CIDRBlock, subnet.CIDRBlock)
}
