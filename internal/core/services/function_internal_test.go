package services

import (
	"archive/zip"
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestFunctionService_InternalExtract(t *testing.T) {
	s := &FunctionService{logger: slog.Default()}
	tmpDir, _ := os.MkdirTemp("", "test-extract-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("extractZip successful", func(t *testing.T) {
		buf := new(bytes.Buffer)
		zw := zip.NewWriter(buf)
		f, _ := zw.Create("hello.txt")
		_, _ = f.Write([]byte("hello world"))
		_ = zw.Close()

		err := s.extractZip(bytes.NewReader(buf.Bytes()), tmpDir)
		assert.NoError(t, err)

		//nolint:gosec
		content, _ := os.ReadFile(filepath.Join(tmpDir, "hello.txt"))
		assert.Equal(t, "hello world", string(content))
	})

	t.Run("extractZip traversal attempt", func(t *testing.T) {
		buf := new(bytes.Buffer)
		zw := zip.NewWriter(buf)
		// Zip file with relative path attempting traversal
		_, _ = zw.Create("../traversal.txt")
		_ = zw.Close()

		err := s.extractZip(bytes.NewReader(buf.Bytes()), tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid file path")
	})
}

func TestFunctionService_BuildTaskOptions(t *testing.T) {
	s := &FunctionService{}
	f := &domain.Function{
		Runtime: "nodejs20",
		Handler: "index.handler",
		MemoryMB: 256,
	}
	tmpDir := "/tmp/fn-123"
	payload := []byte(`{"foo":"bar"}`)

	opts := s.buildTaskOptions(f, tmpDir, payload)
	assert.Equal(t, "node:20-alpine", opts.Image)
	assert.Equal(t, []string{"node", "index.handler"}, opts.Command)
	assert.Contains(t, opts.Env[0], "PAYLOAD=")
	assert.Equal(t, int64(256), opts.MemoryMB)
}
