package setup

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/internal/repositories/noop"
	"github.com/stretchr/testify/assert"
)

type stubNetworkBackend struct {
	backendType string
	pingErr     error
}

func (s stubNetworkBackend) CreateBridge(_ context.Context, _ string, _ int) error  { return nil }
func (s stubNetworkBackend) DeleteBridge(_ context.Context, _ string) error         { return nil }
func (s stubNetworkBackend) ListBridges(_ context.Context) ([]string, error)        { return []string{}, nil }
func (s stubNetworkBackend) AddPort(_ context.Context, _ string, _ string) error    { return nil }
func (s stubNetworkBackend) DeletePort(_ context.Context, _ string, _ string) error { return nil }
func (s stubNetworkBackend) CreateVXLANTunnel(_ context.Context, _ string, _ int, _ string) error {
	return nil
}
func (s stubNetworkBackend) DeleteVXLANTunnel(_ context.Context, _ string, _ string) error {
	return nil
}
func (s stubNetworkBackend) AddFlowRule(_ context.Context, _ string, _ ports.FlowRule) error {
	return nil
}
func (s stubNetworkBackend) DeleteFlowRule(_ context.Context, _ string, _ string) error { return nil }
func (s stubNetworkBackend) ListFlowRules(_ context.Context, _ string) ([]ports.FlowRule, error) {
	return []ports.FlowRule{}, nil
}
func (s stubNetworkBackend) CreateVethPair(_ context.Context, _ string, _ string) error { return nil }
func (s stubNetworkBackend) AttachVethToBridge(_ context.Context, _ string, _ string) error {
	return nil
}
func (s stubNetworkBackend) DeleteVethPair(_ context.Context, _ string) error { return nil }
func (s stubNetworkBackend) SetVethIP(_ context.Context, _ string, _ string, _ string) error {
	return nil
}
func (s stubNetworkBackend) Ping(_ context.Context) error { return s.pingErr }
func (s stubNetworkBackend) Type() string                 { return s.backendType }

const ovsHealthPath = "/health/ovs"

func TestOVSHealthNoBackend(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &platform.Config{Environment: "test"}
	svcs := &Services{}
	handlers := InitHandlers(svcs, nil, logger)

	router := SetupRouter(cfg, logger, handlers, svcs, nil)
	req := httptest.NewRequest(http.MethodGet, ovsHealthPath, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

func TestOVSHealthNoopBackend(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &platform.Config{Environment: "test"}
	svcs := &Services{}
	handlers := InitHandlers(svcs, nil, logger)
	backend := noop.NewNoopNetworkAdapter(logger)

	router := SetupRouter(cfg, logger, handlers, svcs, backend)
	req := httptest.NewRequest(http.MethodGet, ovsHealthPath, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestOVSHealthUnhealthyBackend(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &platform.Config{Environment: "test"}
	svcs := &Services{}
	handlers := InitHandlers(svcs, nil, logger)
	backend := stubNetworkBackend{backendType: "ovs", pingErr: errors.New("ping failed")}

	router := SetupRouter(cfg, logger, handlers, svcs, backend)
	req := httptest.NewRequest(http.MethodGet, ovsHealthPath, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
}

func TestOVSHealthHealthyBackend(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &platform.Config{Environment: "test"}
	svcs := &Services{}
	handlers := InitHandlers(svcs, nil, logger)
	backend := stubNetworkBackend{backendType: "ovs"}

	router := SetupRouter(cfg, logger, handlers, svcs, backend)
	req := httptest.NewRequest(http.MethodGet, ovsHealthPath, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
