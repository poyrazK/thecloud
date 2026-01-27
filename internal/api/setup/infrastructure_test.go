package setup

import (
	"io"
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/platform"
)

const (
	expectedNonNilBackendMsg = "expected non-nil backend"
	expectedNoopBackendFmt   = "expected noop backend, got %s"
)

func TestInitComputeBackendNoop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &platform.Config{ComputeBackend: "noop"}

	backend, err := InitComputeBackend(cfg, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if backend == nil {
		t.Fatal(expectedNonNilBackendMsg)
	}
	if backend.Type() != "noop" {
		t.Fatalf(expectedNoopBackendFmt, backend.Type())
	}
}

func TestInitStorageBackendVariants(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	noopCfg := &platform.Config{StorageBackend: "noop"}
	noopBackend, err := InitStorageBackend(noopCfg, logger)
	if err != nil {
		t.Fatalf("noop backend error: %v", err)
	}
	if noopBackend.Type() != "noop" {
		t.Fatalf(expectedNoopBackendFmt, noopBackend.Type())
	}

	lvmCfg := &platform.Config{StorageBackend: "lvm", LvmVgName: "vg-test"}
	lvmBackend, err := InitStorageBackend(lvmCfg, logger)
	if err != nil {
		t.Fatalf("lvm backend error: %v", err)
	}
	if lvmBackend.Type() != "lvm" {
		t.Fatalf("expected lvm backend, got %s", lvmBackend.Type())
	}

	defaultCfg := &platform.Config{StorageBackend: "unknown"}
	defaultBackend, err := InitStorageBackend(defaultCfg, logger)
	if err != nil {
		t.Fatalf("default backend error: %v", err)
	}
	if defaultBackend.Type() != "noop" {
		t.Fatalf("expected default noop backend, got %s", defaultBackend.Type())
	}
}

func TestInitNetworkBackendNoopFallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	noopCfg := &platform.Config{NetworkBackend: "noop"}
	backend := InitNetworkBackend(noopCfg, logger)
	if backend == nil {
		t.Fatal(expectedNonNilBackendMsg)
	}
	if backend.Type() != "noop" {
		t.Fatalf(expectedNoopBackendFmt, backend.Type())
	}

	defaultCfg := &platform.Config{NetworkBackend: "ovs"}
	fallbackBackend := InitNetworkBackend(defaultCfg, logger)
	if fallbackBackend == nil {
		t.Fatal(expectedNonNilBackendMsg)
	}
	if fallbackBackend.Type() != "noop" && fallbackBackend.Type() != "ovs" {
		t.Fatalf("unexpected backend type %s", fallbackBackend.Type())
	}
}
