//go:build integration
// +build integration

package services_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/core/services"
	"github.com/poyrazk/thecloud/internal/repositories/docker"
	"github.com/poyrazk/thecloud/internal/repositories/filesystem"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const indexJSMockFile = "index.js"

func setupFunctionServiceTest(t *testing.T) (*services.FunctionService, ports.FunctionRepository, ports.SecretService, context.Context) {
	t.Helper()
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewFunctionRepository(db)
	secretRepo := postgres.NewSecretRepository(db)

	compute, err := docker.NewDockerAdapter(slog.Default())
	require.NoError(t, err)

	tmpStorage := t.TempDir()
	fileStore, err := filesystem.NewLocalFileStore(tmpStorage)
	require.NoError(t, err)

	rbacSvc := new(MockRBACService)
	rbacSvc.On("Authorize", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(services.AuditServiceParams{
		Repo:    auditRepo,
		RBACSvc: rbacSvc,
	})

	eventRepo := postgres.NewEventRepository(db)
	eventSvc := services.NewEventService(services.EventServiceParams{
		Repo:    eventRepo,
		RBACSvc: rbacSvc,
	})

	masterKey := "test-master-key-32-chars-long-!!!"
	secretSvc, err := services.NewSecretService(services.SecretServiceParams{
		Repo:        secretRepo,
		RBACSvc:     rbacSvc,
		EventSvc:    eventSvc,
		AuditSvc:    auditSvc,
		Logger:      slog.Default(),
		MasterKey:   masterKey,
		Environment: "test",
	})
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewFunctionService(repo, rbacSvc, compute, fileStore, auditSvc, secretSvc, logger)

	return svc, repo, secretSvc, ctx
}


func createZip(t *testing.T, content string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(indexJSMockFile)
	require.NoError(t, err)
	_, err = f.Write([]byte(content))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)
	return buf.Bytes()
}

func TestFunctionServiceCreateFunctionSuccess(t *testing.T) {
	svc, repo, _, ctx := setupFunctionServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	name := "test-func"
	runtime := "nodejs20"
	handler := indexJSMockFile
	code := createZip(t, "console.log('hello world')")

	f, err := svc.CreateFunction(ctx, name, runtime, handler, code)

	require.NoError(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, name, f.Name)
	assert.Equal(t, "ACTIVE", f.Status)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, f.ID)
	require.NoError(t, err)
	assert.Equal(t, f.ID, fetched.ID)
	assert.Equal(t, userID, fetched.UserID)
}

func TestFunctionServiceInvokeFunctionSuccess(t *testing.T) {
	// Skip if we don't want to actually run docker in all environments,
	// but here we are aiming for real integration.
	svc, _, _, ctx := setupFunctionServiceTest(t)

	code := createZip(t, `
const payload = process.env.PAYLOAD;
console.log("Input: " + payload);
process.exit(0);
`)
	f, err := svc.CreateFunction(ctx, "invoke-test", "nodejs20", indexJSMockFile, code)
	require.NoError(t, err)

	inv, err := svc.InvokeFunction(ctx, f.ID, []byte("hello-from-test"), false)

	require.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, "SUCCESS", inv.Status)
	assert.Contains(t, inv.Logs, "Input: hello-from-test")
}

func TestFunctionServiceDeleteFunctionSuccess(t *testing.T) {
	svc, repo, _, ctx := setupFunctionServiceTest(t)

	code := createZip(t, "console.log(1)")
	f, _ := svc.CreateFunction(ctx, "to-delete", "nodejs20", indexJSMockFile, code)

	err := svc.DeleteFunction(ctx, f.ID)
	require.NoError(t, err)

	// Verify deleted from DB
	_, err = repo.GetByID(ctx, f.ID)
	require.Error(t, err)
}

func TestFunctionServiceListFunctions(t *testing.T) {
	svc, _, _, ctx := setupFunctionServiceTest(t)
	code := createZip(t, "1")
	_, _ = svc.CreateFunction(ctx, "fn1", "nodejs20", indexJSMockFile, code)
	_, _ = svc.CreateFunction(ctx, "fn2", "nodejs20", indexJSMockFile, code)

	fns, err := svc.ListFunctions(ctx)
	require.NoError(t, err)
	assert.Len(t, fns, 2)
}

func TestFunctionServiceGetFunction(t *testing.T) {
	svc, _, _, ctx := setupFunctionServiceTest(t)
	code := createZip(t, "1")
	f, _ := svc.CreateFunction(ctx, "get-me", "nodejs20", indexJSMockFile, code)

	res, err := svc.GetFunction(ctx, f.ID)
	require.NoError(t, err)
	assert.Equal(t, f.ID, res.ID)
}

func TestFunctionServiceInvokeAsync(t *testing.T) {
	svc, repo, _, ctx := setupFunctionServiceTest(t)
	code := createZip(t, "console.log('async')")
	f, _ := svc.CreateFunction(ctx, "async-test", "nodejs20", indexJSMockFile, code)

	inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), true)
	require.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, "PENDING", inv.Status)

	// Wait for async execution and DB record
	var lastInv *domain.Invocation
	require.Eventually(t, func() bool {
		invs, _ := repo.GetInvocations(ctx, f.ID, 1)
		if len(invs) > 0 {
			lastInv = invs[0]
			return lastInv.Status == "SUCCESS"
		}
		return false
	}, 10*time.Second, 500*time.Millisecond)

	assert.Equal(t, "SUCCESS", lastInv.Status)
}

func TestFunctionServiceZipSlipProtection(t *testing.T) {
	svc, _, _, ctx := setupFunctionServiceTest(t)

	// Create malicious zip
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	_, _ = w.Create("../evil.txt")
	_ = w.Close()

	f, err := svc.CreateFunction(ctx, "zip-slip", "nodejs20", indexJSMockFile, buf.Bytes())
	require.NoError(t, err)

	inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), false)
	require.Error(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, "FAILED", inv.Status)
	assert.Contains(t, inv.Logs, "invalid file path in zip")
}

func TestFunctionServiceSecretEnvVarIntegration(t *testing.T) {
	svc, _, secretSvc, ctx := setupFunctionServiceTest(t)

	// Create a secret to reference (resolved by name in sub-tests)
	_, err := secretSvc.CreateSecret(ctx, "my-api-key", "s3cr3t-value", "test secret")
	require.NoError(t, err)

	t.Run("plainTextEnvVar", func(t *testing.T) {
		code := createZip(t, `
console.log("ENV_FOO=" + process.env.FOO);
process.exit(0);
`)
		f, err := svc.CreateFunction(ctx, "env-plain", "nodejs20", indexJSMockFile, code)
		require.NoError(t, err)

		// Update with plain text env var
		timeout := 300
		mem := 256
		_, err = svc.UpdateFunction(ctx, f.ID, &domain.FunctionUpdate{
			EnvVars:  []*domain.EnvVar{{Key: "FOO", Value: "bar"}},
			Timeout:  &timeout,
			MemoryMB: &mem,
		})
		require.NoError(t, err)

		// Invoke and verify
		inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), false)
		require.NoError(t, err)
		assert.Equal(t, "SUCCESS", inv.Status)
		assert.Contains(t, inv.Logs, "ENV_FOO=bar")
	})

	t.Run("secretRefEnvVar", func(t *testing.T) {
		code := createZip(t, `
const foo = JSON.parse(process.env.FOO || '{}');
console.log("KEY=" + foo.key);
console.log("VAL=" + foo.value);
process.exit(0);
`)
		f, err := svc.CreateFunction(ctx, "env-secret", "nodejs20", indexJSMockFile, code)
		require.NoError(t, err)

		// Update with secret ref
		timeout := 300
		mem := 256
		_, err = svc.UpdateFunction(ctx, f.ID, &domain.FunctionUpdate{
			EnvVars:  []*domain.EnvVar{{Key: "FOO", SecretRef: "@my-api-key"}},
			Timeout:  &timeout,
			MemoryMB: &mem,
		})
		require.NoError(t, err)

		// Invoke and verify JSON injection format
		inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), false)
		require.NoError(t, err)
		assert.Equal(t, "SUCCESS", inv.Status)
		assert.Contains(t, inv.Logs, `"key":"FOO"`)
		assert.Contains(t, inv.Logs, `"value":"s3cr3t-value"`)
	})

	t.Run("secretRefMissing_skipsEnvVar", func(t *testing.T) {
		code := createZip(t, `
console.log("HAS_BAZ=" + (process.env.BAZ !== undefined));
process.exit(0);
`)
		f, err := svc.CreateFunction(ctx, "env-missing-secret", "nodejs20", indexJSMockFile, code)
		require.NoError(t, err)

		// Update with reference to non-existent secret
		timeout := 300
		mem := 256
		_, err = svc.UpdateFunction(ctx, f.ID, &domain.FunctionUpdate{
			EnvVars:  []*domain.EnvVar{{Key: "BAZ", SecretRef: "@nonexistent-secret"}},
			Timeout:  &timeout,
			MemoryMB: &mem,
		})
		require.NoError(t, err)

		// Invoke — missing secret should be skipped, BAZ not set
		inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), false)
		require.NoError(t, err)
		assert.Equal(t, "SUCCESS", inv.Status)
		assert.Contains(t, inv.Logs, "HAS_BAZ=false")
	})

	t.Run("mixedPlainTextAndSecretRef", func(t *testing.T) {
		code := createZip(t, `
const foo = JSON.parse(process.env.FOO || '{}');
const bar = process.env.BAR;
console.log("KEY=" + foo.key);
console.log("VAL=" + foo.value);
console.log("BAR=" + bar);
process.exit(0);
`)
		f, err := svc.CreateFunction(ctx, "env-mixed", "nodejs20", indexJSMockFile, code)
		require.NoError(t, err)

		timeout := 300
		mem := 256
		_, err = svc.UpdateFunction(ctx, f.ID, &domain.FunctionUpdate{
			EnvVars: []*domain.EnvVar{
				{Key: "FOO", SecretRef: "@my-api-key"},
				{Key: "BAR", Value: "plain-value"},
			},
			Timeout:  &timeout,
			MemoryMB: &mem,
		})
		require.NoError(t, err)

		inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), false)
		require.NoError(t, err)
		assert.Equal(t, "SUCCESS", inv.Status)
		assert.Contains(t, inv.Logs, `"key":"FOO"`)
		assert.Contains(t, inv.Logs, `"value":"s3cr3t-value"`)
		assert.Contains(t, inv.Logs, "BAR=plain-value")
	})
}
