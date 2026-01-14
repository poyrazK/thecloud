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

func TestClientCreateSecurityGroup(t *testing.T) {
	vpcID := "vpc-123"
	expectedSG := SecurityGroup{
		ID:          "sg-1",
		VPCID:       vpcID,
		Name:        testSgName,
		Description: "test security group",
		CreatedAt:   time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, sgBasePath, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, vpcID, req["vpc_id"])
		assert.Equal(t, expectedSG.Name, req["name"])

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		resp := Response[SecurityGroup]{Data: expectedSG}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	sg, err := client.CreateSecurityGroup(vpcID, testSgName, "test security group")

	assert.NoError(t, err)
	assert.NotNil(t, sg)
	assert.Equal(t, expectedSG.ID, sg.ID)
}

func TestClientListSecurityGroups(t *testing.T) {
	vpcID := "vpc-123"
	expectedSGs := []SecurityGroup{
		{ID: "sg-1", Name: "sg-1", VPCID: vpcID},
		{ID: "sg-2", Name: "sg-2", VPCID: vpcID},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, sgBasePath, r.URL.Path)
		assert.Equal(t, "vpc_id="+vpcID, r.URL.RawQuery)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		resp := Response[[]SecurityGroup]{Data: expectedSGs}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	sgs, err := client.ListSecurityGroups(vpcID)

	assert.NoError(t, err)
	assert.Len(t, sgs, 2)
}

func TestClientGetSecurityGroup(t *testing.T) {
	id := sg123
	expectedSG := SecurityGroup{
		ID:   id,
		Name: testSgName,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, sgDetailPath+id, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		resp := Response[SecurityGroup]{Data: expectedSG}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	sg, err := client.GetSecurityGroup(id)

	assert.NoError(t, err)
	assert.NotNil(t, sg)
	assert.Equal(t, expectedSG.ID, sg.ID)
}

func TestClientDeleteSecurityGroup(t *testing.T) {
	id := sg123

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, sgDetailPath+id, r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.DeleteSecurityGroup(id)

	assert.NoError(t, err)
}

func TestClientAddSecurityRule(t *testing.T) {
	groupID := sg123
	rule := SecurityRule{
		Direction: "ingress",
		Protocol:  "tcp",
		PortMin:   80,
		PortMax:   80,
		CIDR:      testutil.TestAnyCIDR,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, sgDetailPath+groupID+"/rules", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req SecurityRule
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, rule.Protocol, req.Protocol)

		w.Header().Set(contentType, testutil.TestContentTypeAppJSON)
		rule.ID = "rule-1"
		resp := Response[SecurityRule]{Data: rule}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	result, err := client.AddSecurityRule(groupID, rule)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "rule-1", result.ID)
}

func TestClientAttachSecurityGroup(t *testing.T) {
	instanceID := "inst-123"
	groupID := sg123

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/security-groups/attach", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, instanceID, req["instance_id"])
		assert.Equal(t, groupID, req["group_id"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, testAPIKey)
	err := client.AttachSecurityGroup(instanceID, groupID)

	assert.NoError(t, err)
}
