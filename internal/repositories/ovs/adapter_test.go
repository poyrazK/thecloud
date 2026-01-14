package ovs_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/repositories/ovs"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSrcFmt        = "ip,nw_src=%s"
	ovsNotAvailable   = "OVS not available, skipping validation test"
	errInvalidBridge  = "expected error for invalid bridge name"
	invalidBridgeName = "invalid bridge"
)

func TestOvsAdapterIntegration(t *testing.T) {
	if os.Getenv("OVS_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping OVS integration test. Set OVS_INTEGRATION_TEST=true to run.")
	}

	if _, err := os.Stat("/usr/bin/ovs-vsctl"); os.IsNotExist(err) {
		t.Skip("Skipping OVS integration test: ovs-vsctl not found at /usr/bin/ovs-vsctl")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	require.NoError(t, err)

	ctx := context.Background()
	bridgeName := "br-test-" + uuid.New().String()[:8]

	// 1. Create Bridge
	t.Run("CreateBridge", func(t *testing.T) {
		err := adapter.CreateBridge(ctx, bridgeName, 100)
		assert.NoError(t, err)
	})

	// 2. Add Flow Rule
	t.Run("AddFlowRule", func(t *testing.T) {
		rule := ports.FlowRule{
			Priority: 100,
			Match:    fmt.Sprintf(testSrcFmt, testutil.TestIPHost),
			Actions:  "drop",
		}
		err := adapter.AddFlowRule(ctx, bridgeName, rule)
		assert.NoError(t, err)
	})

	// 3. Delete Flow Rule
	t.Run("DeleteFlowRule", func(t *testing.T) {
		err := adapter.DeleteFlowRule(ctx, bridgeName, fmt.Sprintf(testSrcFmt, testutil.TestIPHost))
		assert.NoError(t, err)
	})

	// 4. Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		err := adapter.DeleteBridge(ctx, bridgeName)
		assert.NoError(t, err)
	})
}

func TestOvsAdapterType(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip("OVS not available, skipping type test")
	}

	if adapter.Type() != "ovs" {
		t.Fatalf("expected type 'ovs', got %s", adapter.Type())
	}
}

func TestOvsAdapterCreateBridgeInvalidName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip(ovsNotAvailable)
	}

	err = adapter.CreateBridge(context.Background(), "invalid name", 1)
	if err == nil {
		t.Fatal(errInvalidBridge)
	}
}

func TestOvsAdapterDeleteBridgeInvalidName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip(ovsNotAvailable)
	}

	err = adapter.DeleteBridge(context.Background(), "invalid name")
	if err == nil {
		t.Fatal(errInvalidBridge)
	}
}

func TestOvsAdapterAddPortInvalidName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip(ovsNotAvailable)
	}

	err = adapter.AddPort(context.Background(), invalidBridgeName, "port")
	if err == nil {
		t.Fatal(errInvalidBridge)
	}
}

func TestOvsAdapterDeletePortInvalidName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip(ovsNotAvailable)
	}

	err = adapter.DeletePort(context.Background(), "bridge", "invalid port")
	if err == nil {
		t.Fatalf("expected error for invalid port name")
	}
}

func TestOvsAdapterAddFlowRuleInvalidBridge(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip(ovsNotAvailable)
	}

	rule := ports.FlowRule{Priority: 100, Match: "in_port=1", Actions: "output:2"}
	err = adapter.AddFlowRule(context.Background(), invalidBridgeName, rule)
	if err == nil {
		t.Fatal(errInvalidBridge)
	}
}

func TestOvsAdapterDeleteFlowRuleInvalidBridge(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip(ovsNotAvailable)
	}

	err = adapter.DeleteFlowRule(context.Background(), invalidBridgeName, "match")
	if err == nil {
		t.Fatal(errInvalidBridge)
	}
}

func TestOvsAdapterListFlowRules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	adapter, err := ovs.NewOvsAdapter(logger)
	if err != nil {
		t.Skip("OVS not available, skipping list test")
	}

	rules, err := adapter.ListFlowRules(context.Background(), "bridge")
	if err != nil {
		t.Fatalf("ListFlowRules failed: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected empty rules, got %d", len(rules))
	}
}
