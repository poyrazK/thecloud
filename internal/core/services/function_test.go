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
	"github.com/stretchr/testify/require"
)

func setupFunctionServiceTest(t *testing.T) (*services.FunctionService, ports.FunctionRepository, ports.ComputeBackend, ports.FileStore, context.Context) {
	db := setupDB(t)
	cleanDB(t, db)
	ctx := setupTestUser(t, db)

	repo := postgres.NewFunctionRepository(db)

	compute, err := docker.NewDockerAdapter(slog.Default())
	require.NoError(t, err)

	tmpStorage := t.TempDir()
	fileStore, err := filesystem.NewLocalFileStore(tmpStorage)
	require.NoError(t, err)

	auditRepo := postgres.NewAuditRepository(db)
	auditSvc := services.NewAuditService(auditRepo)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := services.NewFunctionService(repo, compute, fileStore, auditSvc, logger)
	return svc, repo, compute, fileStore, ctx
}

func createZip(t *testing.T, filename, content string) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(filename)
	require.NoError(t, err)
	_, err = f.Write([]byte(content))
	require.NoError(t, err)
	err = w.Close()
	require.NoError(t, err)
	return buf.Bytes()
}

func TestFunctionService_CreateFunction_Success(t *testing.T) {
	svc, repo, _, _, ctx := setupFunctionServiceTest(t)
	userID := appcontext.UserIDFromContext(ctx)

	name := "test-func"
	runtime := "nodejs20"
	handler := "index.js"
	code := createZip(t, "index.js", "console.log('hello world')")

	f, err := svc.CreateFunction(ctx, name, runtime, handler, code)

	assert.NoError(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, name, f.Name)
	assert.Equal(t, "ACTIVE", f.Status)

	// Verify in DB
	fetched, err := repo.GetByID(ctx, f.ID)
	assert.NoError(t, err)
	assert.Equal(t, f.ID, fetched.ID)
	assert.Equal(t, userID, fetched.UserID)
}

func TestFunctionService_InvokeFunction_Success(t *testing.T) {
	// Skip if we don't want to actually run docker in all environments,
	// but here we are aiming for real integration.
	svc, _, _, _, ctx := setupFunctionServiceTest(t)

	code := createZip(t, "index.js", `
const payload = process.env.PAYLOAD;
console.log("Input: " + payload);
process.exit(0);
`)
	f, err := svc.CreateFunction(ctx, "invoke-test", "nodejs20", "index.js", code)
	require.NoError(t, err)

	inv, err := svc.InvokeFunction(ctx, f.ID, []byte("hello-from-test"), false)

	assert.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, "SUCCESS", inv.Status)
	assert.Contains(t, inv.Logs, "Input: hello-from-test")
}

func TestFunctionService_DeleteFunction_Success(t *testing.T) {
	svc, repo, _, _, ctx := setupFunctionServiceTest(t)

	code := createZip(t, "index.js", "console.log(1)")
	f, _ := svc.CreateFunction(ctx, "to-delete", "nodejs20", "index.js", code)

	err := svc.DeleteFunction(ctx, f.ID)
	assert.NoError(t, err)

	// Verify deleted from DB
	_, err = repo.GetByID(ctx, f.ID)
	assert.Error(t, err)
}

func TestFunctionService_ListFunctions(t *testing.T) {
	svc, _, _, _, ctx := setupFunctionServiceTest(t)
	code := createZip(t, "index.js", "1")
	_, _ = svc.CreateFunction(ctx, "fn1", "nodejs20", "index.js", code)
	_, _ = svc.CreateFunction(ctx, "fn2", "nodejs20", "index.js", code)

	fns, err := svc.ListFunctions(ctx)
	assert.NoError(t, err)
	assert.Len(t, fns, 2)
}

func TestFunctionService_GetFunction(t *testing.T) {
	svc, _, _, _, ctx := setupFunctionServiceTest(t)
	code := createZip(t, "index.js", "1")
	f, _ := svc.CreateFunction(ctx, "get-me", "nodejs20", "index.js", code)

	res, err := svc.GetFunction(ctx, f.ID)
	assert.NoError(t, err)
	assert.Equal(t, f.ID, res.ID)
}

func TestFunctionService_InvokeAsync(t *testing.T) {
	svc, repo, _, _, ctx := setupFunctionServiceTest(t)
	code := createZip(t, "index.js", "console.log('async')")
	f, _ := svc.CreateFunction(ctx, "async-test", "nodejs20", "index.js", code)

	inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), true)
	assert.NoError(t, err)
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

func TestFunctionService_ZipSlipProtection(t *testing.T) {
	svc, _, _, _, ctx := setupFunctionServiceTest(t)

	// Create malicious zip
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	_, _ = w.Create("../evil.txt")
	_ = w.Close()

	f, err := svc.CreateFunction(ctx, "zip-slip", "nodejs20", "index.js", buf.Bytes())
	require.NoError(t, err)

	inv, err := svc.InvokeFunction(ctx, f.ID, []byte("{}"), false)
	assert.Error(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, "FAILED", inv.Status)
	assert.Contains(t, inv.Logs, "invalid file path in zip")
}
