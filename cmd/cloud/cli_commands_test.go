package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/sdk"
)

const (
	testSGID      = "sg-1"
	testVPCID     = "vpc-1"
	testCacheID   = "cache-1"
	testQueueID   = "queue-1"
	testTopicID   = "topic-1"
	testRedisMain = "redis-main"
	testJobs      = "jobs"
	testUpdates   = "updates"
	errSuccessMsg = "expected success message, got: %s"
	testAPIKeyStr = "api-key"
	pathSG        = "/security-groups/"
	pathVPC       = "/vpcs/"
	pathCache     = "/caches/"
	pathQueue     = "/queues/"
	pathDB        = "/databases/"
	pathSecret    = "/secrets/"
	pathDBConn    = "/connection"
	pathStats     = "/stats"
	pathFlush     = "/flush"
	pathMessages  = "/messages"
	pathPurge     = "/purge"
	pathRules     = "/rules"
)

func setupAPIServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path
		method := r.Method

		handled := handleSecurityGroupsMock(w, r, method, path) ||
			handleVPCsMock(w, r, method, path) ||
			handleVolumesMock(w, r, method, path) ||
			handleCachesMock(w, r, method, path) ||
			handleQueuesMock(w, r, method, path) ||
			handleNotifyMock(w, r, method, path) ||
			handleRBACMock(w, r, method, path) ||
			handleDatabasesMock(w, r, method, path) ||
			handleSecretsMock(w, r, method, path)

		if !handled {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func handleSecurityGroupsMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodGet && path == "/security-groups":
		resp := sdk.Response[[]sdk.SecurityGroup]{
			Data: []sdk.SecurityGroup{{ID: testSGID, VPCID: testVPCID, Name: "default", Description: "default group", ARN: "arn:sg:1", CreatedAt: time.Now().UTC()}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/security-groups":
		resp := sdk.Response[sdk.SecurityGroup]{
			Data: sdk.SecurityGroup{ID: "sg-2", VPCID: "vpc-2", Name: "web", Description: "web group", ARN: "arn:sg:2", CreatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == pathSG+testSGID:
		resp := sdk.Response[sdk.SecurityGroup]{
			Data: sdk.SecurityGroup{ID: testSGID, VPCID: testVPCID, Name: "default", Description: "default group", ARN: "arn:sg:1", CreatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && path == pathSG+testSGID:
		return respondNoContent(w)
	case method == http.MethodPost && strings.HasPrefix(path, pathSG) && strings.HasSuffix(path, pathRules):
		resp := sdk.Response[sdk.SecurityRule]{
			Data: sdk.SecurityRule{ID: "rule-1", Direction: "ingress", Protocol: "tcp", PortMin: 80, PortMax: 80, CIDR: "0.0.0.0/0", Priority: 100},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && strings.HasPrefix(path, pathSG+"rules/"):
		return respondNoContent(w)
	case (method == http.MethodPost && path == "/security-groups/attach") ||
		(method == http.MethodPost && path == "/security-groups/detach"):
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func handleVPCsMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodGet && path == "/vpcs":
		resp := sdk.Response[[]sdk.VPC]{
			Data: []sdk.VPC{{ID: testVPCID, Name: "main", CIDRBlock: "10.0.0.0/16", NetworkID: "net-1", VXLANID: 1001, Status: "available", CreatedAt: time.Now().UTC()}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/vpcs":
		resp := sdk.Response[sdk.VPC]{
			Data: sdk.VPC{ID: "vpc-2", Name: "demo", CIDRBlock: "10.1.0.0/16", NetworkID: "net-2", VXLANID: 1002, Status: "available", CreatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == pathVPC+testVPCID+"/subnets":
		resp := sdk.Response[[]*sdk.Subnet]{
			Data: []*sdk.Subnet{{ID: "subnet-1", VpcID: testVPCID, Name: "public", CIDRBlock: "10.0.1.0/24", AZ: "us-east-1a", GatewayIP: "10.0.1.1", Status: "available", CreatedAt: time.Now().UTC()}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == pathVPC+testVPCID+"/subnets":
		resp := sdk.Response[*sdk.Subnet]{
			Data: &sdk.Subnet{ID: "subnet-2", VpcID: testVPCID, Name: "private", CIDRBlock: "10.0.2.0/24", AZ: "us-east-1b", GatewayIP: "10.0.2.1", Status: "available", CreatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case (method == http.MethodDelete && path == pathVPC+testVPCID) ||
		(method == http.MethodDelete && path == "/subnets/subnet-1"):
		return respondNoContent(w)
	}
	return false
}

func handleVolumesMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodGet && path == "/volumes":
		resp := sdk.Response[[]sdk.Volume]{
			Data: []sdk.Volume{{ID: uuid.New(), Name: "data", SizeGB: 20, Status: "available", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/volumes":
		resp := sdk.Response[sdk.Volume]{
			Data: sdk.Volume{ID: uuid.New(), Name: "data", SizeGB: 20, Status: "available", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && path == "/volumes/vol-1":
		return respondNoContent(w)
	}
	return false
}

func handleCachesMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	vpcID := testVPCID
	cacheID := testCacheID // Assuming testCacheID is now a local var or defined elsewhere
	switch {
	case method == http.MethodPost && path == "/caches":
		resp := sdk.Response[sdk.Cache]{
			Data: sdk.Cache{ID: cacheID, Name: testRedisMain, Engine: "redis", Version: "7.2", Status: "PROVISIONING", VpcID: &vpcID, Port: 6379, MemoryMB: 128, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == "/caches":
		resp := sdk.Response[[]*sdk.Cache]{
			Data: []*sdk.Cache{{ID: cacheID, Name: testRedisMain, Engine: "redis", Version: "7.2", Status: "RUNNING", VpcID: &vpcID, Port: 6379, MemoryMB: 128}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == pathCache+cacheID:
		resp := sdk.Response[sdk.Cache]{
			Data: sdk.Cache{ID: cacheID, Name: testRedisMain, Engine: "redis", Version: "7.2", Status: "RUNNING", VpcID: &vpcID, Port: 6379, MemoryMB: 128},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && path == pathCache+cacheID:
		return respondNoContent(w)
	case method == http.MethodGet && path == pathCache+cacheID+pathDBConn:
		resp := sdk.Response[map[string]string]{
			Data: map[string]string{"connection_string": "redis://" + cacheID + ":6379"},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == pathCache+cacheID+pathFlush:
		return respondNoContent(w)
	case method == http.MethodGet && path == pathCache+cacheID+pathStats:
		resp := sdk.Response[sdk.CacheStats]{
			Data: sdk.CacheStats{UsedMemoryBytes: 1024, MaxMemoryBytes: 2048, ConnectedClients: 5, TotalKeys: 10},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	}
	return false
}

func handleQueuesMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodPost && path == "/queues":
		resp := sdk.Response[sdk.Queue]{
			Data: sdk.Queue{ID: testQueueID, Name: testJobs, ARN: "arn:queue:1", Status: "ACTIVE"},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == "/queues":
		resp := sdk.Response[[]sdk.Queue]{
			Data: []sdk.Queue{{ID: testQueueID, Name: testJobs, ARN: "arn:queue:1", Status: "ACTIVE"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && path == pathQueue+testQueueID:
		return respondNoContent(w)
	case method == http.MethodPost && path == pathQueue+testQueueID+pathMessages:
		resp := sdk.Response[sdk.Message]{
			Data: sdk.Message{ID: "msg-1", QueueID: testQueueID, Body: "hello", ReceiptHandle: "rh-1", CreatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == pathQueue+testQueueID+pathMessages:
		resp := sdk.Response[[]sdk.Message]{
			Data: []sdk.Message{{ID: "msg-1", QueueID: testQueueID, Body: "hello", ReceiptHandle: "rh-1", CreatedAt: time.Now().UTC()}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case (method == http.MethodDelete && strings.HasPrefix(path, pathQueue+testQueueID+pathMessages+"/")) ||
		(method == http.MethodPost && path == pathQueue+testQueueID+pathPurge):
		return respondNoContent(w)
	}
	return false
}

func handleNotifyMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodPost && path == "/notify/topics":
		resp := sdk.Response[sdk.Topic]{
			Data: sdk.Topic{ID: testTopicID, Name: testUpdates, ARN: "arn:topic:1"},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == "/notify/topics":
		resp := sdk.Response[[]sdk.Topic]{
			Data: []sdk.Topic{{ID: testTopicID, Name: testUpdates, ARN: "arn:topic:1"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/notify/topics/"+testTopicID+"/subscriptions":
		resp := sdk.Response[sdk.Subscription]{
			Data: sdk.Subscription{ID: "sub-1", TopicID: testTopicID, Protocol: "webhook", Endpoint: "https://example.com"},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/notify/topics/"+testTopicID+"/publish":
		return respondNoContent(w)
	}
	return false
}

func handleRBACMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodGet && path == "/rbac/roles":
		resp := sdk.Response[[]domain.Role]{
			Data: []domain.Role{{ID: uuid.New(), Name: "admin", Permissions: []domain.Permission{"*"}}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/rbac/roles":
		resp := sdk.Response[domain.Role]{
			Data: domain.Role{ID: uuid.New(), Name: "editor", Permissions: []domain.Permission{"read", "write"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && strings.HasPrefix(path, "/rbac/roles/"):
		w.WriteHeader(http.StatusNoContent)
		return true
	case method == http.MethodGet && path == "/rbac/bindings":
		resp := sdk.Response[[]domain.User]{
			Data: []domain.User{{ID: uuid.New(), Email: "user@example.com", Role: "admin"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/rbac/bindings":
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func handleDatabasesMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodGet && path == "/databases":
		resp := sdk.Response[[]sdk.Database]{
			Data: []sdk.Database{{ID: "db-1", Name: "main", Engine: "postgres", Status: "running", Port: 5432}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/databases":
		resp := sdk.Response[sdk.Database]{
			Data: sdk.Database{ID: "db-1", Name: "main", Username: "admin", Password: "pwd", Port: 5432},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == pathDB+"db-1":
		resp := sdk.Response[sdk.Database]{
			Data: sdk.Database{ID: "db-1", Name: "main", Engine: "postgres", Status: "running", Port: 5432},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && path == pathDB+"db-1":
		return respondNoContent(w)
	case method == http.MethodGet && path == pathDB+"db-1"+pathDBConn:
		resp := sdk.Response[map[string]string]{Data: map[string]string{"connection_string": "postgres://admin@host:5432/main"}}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	}
	return false
}

func handleSecretsMock(w http.ResponseWriter, _ *http.Request, method, path string) bool {
	switch {
	case method == http.MethodGet && path == "/secrets":
		resp := sdk.Response[[]sdk.Secret]{
			Data: []sdk.Secret{{ID: "sec-1", Name: testAPIKeyStr, Description: "prod key", CreatedAt: time.Now().UTC()}},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodPost && path == "/secrets":
		resp := sdk.Response[sdk.Secret]{
			Data: sdk.Secret{ID: "sec-1", Name: testAPIKeyStr, CreatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodGet && path == pathSecret+"sec-1":
		resp := sdk.Response[sdk.Secret]{
			Data: sdk.Secret{ID: "sec-1", Name: testAPIKeyStr, EncryptedValue: "decrypted-value", CreatedAt: time.Now().UTC()},
		}
		_ = json.NewEncoder(w).Encode(resp)
		return true
	case method == http.MethodDelete && path == pathSecret+"sec-1":
		return respondNoContent(w)
	}
	return false
}

func setAPIContext(t *testing.T, server *httptest.Server) {
	t.Helper()

	oldURL := apiURL
	oldKey := apiKey
	oldJSON := outputJSON

	apiURL = server.URL
	apiKey = "test-key"
	outputJSON = false

	t.Cleanup(func() {
		apiURL = oldURL
		apiKey = oldKey
		outputJSON = oldJSON
	})
}

func TestSGListCommandJSONOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = sgListCmd.Flags().Set(flagVPCID, "vpc-1")
	outputJSON = true
	t.Cleanup(func() {
		_ = sgListCmd.Flags().Set(flagVPCID, "")
		outputJSON = false
	})

	out := captureStdout(t, func() {
		sgListCmd.Run(sgListCmd, nil)
	})
	if !strings.Contains(out, "\"id\": \"sg-1\"") {
		t.Fatalf("expected security group in output, got: %s", out)
	}
}

func TestSGCreateCommandMissingVPC(t *testing.T) {
	out := captureStdout(t, func() {
		sgCreateCmd.Run(sgCreateCmd, []string{"web"})
	})
	if !strings.Contains(out, "--vpc-id is required") {
		t.Fatalf("expected missing vpc-id error, got: %s", out)
	}
}

func TestVPCCreateCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		vpcCreateCmd.Run(vpcCreateCmd, []string{"demo"})
	})
	if !strings.Contains(out, "VPC demo created successfully") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestVPCListCommandJSONOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	outputJSON = true
	t.Cleanup(func() { outputJSON = false })

	out := captureStdout(t, func() {
		vpcListCmd.Run(vpcListCmd, nil)
	})
	if !strings.Contains(out, "\"id\": \"vpc-1\"") {
		t.Fatalf("expected VPC in output, got: %s", out)
	}
}

func TestRBACRoleCreateCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		createRoleCmd.Run(createRoleCmd, []string{"editor"})
	})
	if !strings.Contains(out, "[SUCCESS] Role created: editor") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestRBACRoleListCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		listRolesCmd.Run(listRolesCmd, nil)
	})
	if !strings.Contains(out, "admin") {
		t.Fatalf("expected role in output, got: %s", out)
	}
}

func TestRBACRoleBindCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		bindRoleCmd.Run(bindRoleCmd, []string{"user-1", "admin"})
	})
	if !strings.Contains(out, "[SUCCESS] Role admin bound to user user-1") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestRBACRoleDeleteCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	id := uuid.New()
	out := captureStdout(t, func() {
		deleteRoleCmd.Run(deleteRoleCmd, []string{id.String()})
	})
	if !strings.Contains(out, "[SUCCESS] Role "+id.String()+" deleted") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestDBListCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		dbListCmd.Run(dbListCmd, nil)
	})
	if !strings.Contains(out, "main") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestDBCreateCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = dbCreateCmd.Flags().Set("name", "mydb")
	t.Cleanup(func() { _ = dbCreateCmd.Flags().Set("name", "") })

	out := captureStdout(t, func() {
		dbCreateCmd.Run(dbCreateCmd, nil)
	})
	if !strings.Contains(out, "[SUCCESS] Database mydb created successfully!") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestDBShowCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		dbShowCmd.Run(dbShowCmd, []string{"db-1"})
	})
	if !strings.Contains(out, "main") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestDBRmCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		dbRmCmd.Run(dbRmCmd, []string{"db-1"})
	})
	if !strings.Contains(out, "[SUCCESS] Database removed.") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestDBConnCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		dbConnCmd.Run(dbConnCmd, []string{"db-1"})
	})
	if !strings.Contains(out, "postgres://") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSecretsListCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		secretsListCmd.Run(secretsListCmd, nil)
	})
	if !strings.Contains(out, "api-key") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSecretsCreateCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = secretsCreateCmd.Flags().Set("name", "my-secret")
	_ = secretsCreateCmd.Flags().Set("value", "s3cret")
	t.Cleanup(func() {
		_ = secretsCreateCmd.Flags().Set("name", "")
		_ = secretsCreateCmd.Flags().Set("value", "")
	})

	out := captureStdout(t, func() {
		secretsCreateCmd.Run(secretsCreateCmd, nil)
	})
	if !strings.Contains(out, "[SUCCESS] Secret my-secret created.") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSecretsGetCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		secretsGetCmd.Run(secretsGetCmd, []string{"sec-1"})
	})
	if !strings.Contains(out, "decrypted-value") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSecretsRmCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		secretsRmCmd.Run(secretsRmCmd, []string{"sec-1"})
	})
	if !strings.Contains(out, "[SUCCESS] Secret removed.") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSubnetListCommandJSONOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	outputJSON = true
	t.Cleanup(func() { outputJSON = false })

	out := captureStdout(t, func() {
		subnetListCmd.Run(subnetListCmd, []string{"vpc-1"})
	})
	if !strings.Contains(out, "\"id\": \"subnet-1\"") {
		t.Fatalf("expected subnet in output, got: %s", out)
	}
}

func TestVolumeCreateCommandJSONOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = volumeCreateCmd.Flags().Set("name", "data")
	_ = volumeCreateCmd.Flags().Set("size", "20")
	defer func() {
		_ = volumeCreateCmd.Flags().Set("name", "")
		_ = volumeCreateCmd.Flags().Set("size", "1")
	}()

	out := captureStdout(t, func() {
		volumeCreateCmd.Run(volumeCreateCmd, nil)
	})
	if !strings.Contains(out, "Volume created") {
		t.Fatalf("expected volume create output, got: %s", out)
	}
	if !strings.Contains(out, "\"name\": \"data\"") {
		t.Fatalf("expected volume json output, got: %s", out)
	}
}

func TestVolumeListCommandJSONOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	outputJSON = true
	t.Cleanup(func() { outputJSON = false })

	out := captureStdout(t, func() {
		volumeListCmd.Run(volumeListCmd, nil)
	})
	if !strings.Contains(out, "\"name\": \"data\"") {
		t.Fatalf("expected volume list output, got: %s", out)
	}
}

func TestVolumeDeleteCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		volumeDeleteCmd.Run(volumeDeleteCmd, []string{"vol-1"})
	})
	if !strings.Contains(out, "Volume vol-1 deleted") {
		t.Fatalf("expected volume delete output, got: %s", out)
	}
}

func TestVPCRmCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		vpcRmCmd.Run(vpcRmCmd, []string{"vpc-1"})
	})
	if !strings.Contains(out, "VPC vpc-1 removed") {
		t.Fatalf("expected VPC remove output, got: %s", out)
	}
}

func TestSubnetDeleteCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		subnetDeleteCmd.Run(subnetDeleteCmd, []string{"subnet-1"})
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Fatalf("expected subnet delete output, got: %s", out)
	}
}

func TestCacheCreateCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = createCacheCmd.Flags().Set("name", "redis-main")
	_ = createCacheCmd.Flags().Set("version", "7.2")
	_ = createCacheCmd.Flags().Set("memory", "128")
	_ = createCacheCmd.Flags().Set("vpc", "vpc-1")
	_ = createCacheCmd.Flags().Set("wait", "false")
	defer func() {
		_ = createCacheCmd.Flags().Set("name", "")
		_ = createCacheCmd.Flags().Set("version", "7.2")
		_ = createCacheCmd.Flags().Set("memory", "128")
		_ = createCacheCmd.Flags().Set("vpc", "")
		_ = createCacheCmd.Flags().Set("wait", "false")
	}()

	out := captureStdout(t, func() {
		createCacheCmd.Run(createCacheCmd, nil)
	})
	if !strings.Contains(out, "Cache created with ID: cache-1") {
		t.Fatalf("expected cache create output, got: %s", out)
	}
}

func TestCacheListCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		listCacheCmd.Run(listCacheCmd, nil)
	})
	if !strings.Contains(out, testCacheID) {
		t.Fatalf("expected cache list output, got: %s", out)
	}
}

func TestCacheShowCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		getCacheCmd.Run(getCacheCmd, []string{testCacheID})
	})
	if !strings.Contains(out, "Name:      redis-main") {
		t.Fatalf("expected cache show output, got: %s", out)
	}
}

func TestCacheDeleteCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		deleteCacheCmd.Run(deleteCacheCmd, []string{testCacheID})
	})
	if !strings.Contains(out, "Cache deleted successfully") {
		t.Fatalf("expected cache delete output, got: %s", out)
	}
}

func TestCacheConnectionCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		connectionCacheCmd.Run(connectionCacheCmd, []string{testCacheID})
	})
	if !strings.Contains(out, "redis://cache-1:6379") {
		t.Fatalf("expected cache connection output, got: %s", out)
	}
}

func TestCacheStatsCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		statsCacheCmd.Run(statsCacheCmd, []string{testCacheID})
	})
	if !strings.Contains(out, "Used Memory: 1.0 KB") {
		t.Fatalf("expected cache stats output, got: %s", out)
	}
}

func TestCacheFlushCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = flushCacheCmd.Flags().Set("yes", "true")
	t.Cleanup(func() {
		_ = flushCacheCmd.Flags().Set("yes", "false")
	})

	out := captureStdout(t, func() {
		flushCacheCmd.Run(flushCacheCmd, []string{testCacheID})
	})
	if !strings.Contains(out, "Cache flushed successfully") {
		t.Fatalf("expected cache flush output, got: %s", out)
	}
}

func TestQueueListCommandJSONOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	outputJSON = true
	t.Cleanup(func() { outputJSON = false })

	out := captureStdout(t, func() {
		listQueuesCmd.Run(listQueuesCmd, nil)
	})
	if !strings.Contains(out, "\"id\": \"queue-1\"") {
		t.Fatalf("expected queue list output, got: %s", out)
	}
}

func TestQueueCreateCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		createQueueCmd.Run(createQueueCmd, []string{"jobs"})
	})
	if !strings.Contains(out, "Queue created successfully") {
		t.Fatalf("expected queue create output, got: %s", out)
	}
}

func TestQueueDeleteCommandOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		deleteQueueCmd.Run(deleteQueueCmd, []string{testQueueID})
	})
	if !strings.Contains(out, "Queue deleted successfully") {
		t.Fatalf("expected queue delete output, got: %s", out)
	}
}

func TestQueueSendMessageOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		sendMessageCmd.Run(sendMessageCmd, []string{testQueueID, "hello"})
	})
	if !strings.Contains(out, "Message sent") {
		t.Fatalf("expected send message output, got: %s", out)
	}
}

func TestQueueReceiveMessagesJSONOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = receiveMessagesCmd.Flags().Set("max", "1")
	outputJSON = true
	t.Cleanup(func() {
		_ = receiveMessagesCmd.Flags().Set("max", "1")
		outputJSON = false
	})

	out := captureStdout(t, func() {
		receiveMessagesCmd.Run(receiveMessagesCmd, []string{testQueueID})
	})
	if !strings.Contains(out, "\"id\": \"msg-1\"") {
		t.Fatalf("expected receive messages output, got: %s", out)
	}
}

func TestQueueAckMessageOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		ackMessageCmd.Run(ackMessageCmd, []string{testQueueID, "rh-1"})
	})
	if !strings.Contains(out, "Message acknowledged") {
		t.Fatalf("expected ack output, got: %s", out)
	}
}

func TestQueuePurgeOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		purgeQueueCmd.Run(purgeQueueCmd, []string{testQueueID})
	})
	if !strings.Contains(out, "Queue purged") {
		t.Fatalf("expected queue purge output, got: %s", out)
	}
}

func TestNotifyCreateTopicOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		createTopicCmd.Run(createTopicCmd, []string{"updates"})
	})
	if !strings.Contains(out, "Topic created") {
		t.Fatalf("expected create topic output, got: %s", out)
	}
}

func TestNotifyListTopicsOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		listTopicsCmd.Run(listTopicsCmd, nil)
	})
	if !strings.Contains(out, "updates") {
		t.Fatalf("expected list topics output, got: %s", out)
	}
}

func TestNotifySubscribeOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = subscribeCmd.Flags().Set("protocol", "webhook")
	_ = subscribeCmd.Flags().Set("endpoint", "https://example.com")
	defer func() {
		_ = subscribeCmd.Flags().Set("endpoint", "")
	}()

	out := captureStdout(t, func() {
		subscribeCmd.Run(subscribeCmd, []string{testTopicID})
	})
	if !strings.Contains(out, "Subscription created") {
		t.Fatalf("expected subscribe output, got: %s", out)
	}
}

func TestNotifyPublishOutput(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		publishCmd.Run(publishCmd, []string{testTopicID, "hello"})
	})
	if !strings.Contains(out, "Message published") {
		t.Fatalf("expected publish output, got: %s", out)
	}
}

func TestSGGetCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		sgGetCmd.Run(sgGetCmd, []string{"sg-1"})
	})
	if !strings.Contains(out, "Security Group: default") {
		t.Fatalf("expected sg detail output, got: %s", out)
	}
}

func TestSGDeleteCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		sgDeleteCmd.Run(sgDeleteCmd, []string{"sg-1"})
	})
	if !strings.Contains(out, "Security Group sg-1 deleted successfully") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSGAddRuleCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	_ = sgAddRuleCmd.Flags().Set("direction", "ingress")
	_ = sgAddRuleCmd.Flags().Set("protocol", "tcp")
	_ = sgAddRuleCmd.Flags().Set("port-min", "80")
	_ = sgAddRuleCmd.Flags().Set("port-max", "80")
	_ = sgAddRuleCmd.Flags().Set("cidr", "0.0.0.0/0")

	out := captureStdout(t, func() {
		sgAddRuleCmd.Run(sgAddRuleCmd, []string{"sg-1"})
	})
	if !strings.Contains(out, "added successfully to security group sg-1") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSGRemoveRuleCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		sgRemoveRuleCmd.Run(sgRemoveRuleCmd, []string{"rule-1"})
	})
	if !strings.Contains(out, "Rule rule-1 removed successfully") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSGAttachCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		sgAttachCmd.Run(sgAttachCmd, []string{"i-1", "sg-1"})
	})
	if !strings.Contains(out, "attached to instance i-1 successfully") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func TestSGDetachCommandSuccess(t *testing.T) {
	server := setupAPIServer(t)
	defer server.Close()
	setAPIContext(t, server)

	out := captureStdout(t, func() {
		sgDetachCmd.Run(sgDetachCmd, []string{"i-1", "sg-1"})
	})
	if !strings.Contains(out, "detached from instance i-1 successfully") {
		t.Fatalf(errSuccessMsg, out)
	}
}

func respondNoContent(w http.ResponseWriter) bool {
	w.WriteHeader(http.StatusNoContent)
	return true
}
