package ovs

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/poyrazk/thecloud/internal/core/ports"
	apperrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/require"
)

const (
	ovsVsctlPath = "/bin/ovs-vsctl"
	ovsOfctlPath = "/bin/ovs-ofctl"
	badBridge    = "bad bridge"
)

type fakeCmd struct {
	runErr  error
	out     []byte
	outErr  error
	runHits int
}

func (c *fakeCmd) Run() error {
	c.runHits++
	return c.runErr
}

func (c *fakeCmd) Output() ([]byte, error) {
	if c.outErr != nil {
		return nil, c.outErr
	}
	return c.out, nil
}

type fakeExecer struct {
	lookPath map[string]string
	lookErr  error
	cmd      *fakeCmd
}

func (e *fakeExecer) LookPath(file string) (string, error) {
	if e.lookErr != nil {
		return "", e.lookErr
	}
	if p, ok := e.lookPath[file]; ok {
		return p, nil
	}
	return "", errors.New("not found")
}

func (e *fakeExecer) CommandContext(ctx context.Context, name string, args ...string) cmd {
	return e.cmd
}

func TestOvsAdapterCommandErrorsAreWrapped(t *testing.T) {
	fx := &fakeExecer{
		lookPath: map[string]string{"ovs-vsctl": ovsVsctlPath, "ovs-ofctl": ovsOfctlPath},
		cmd:      &fakeCmd{runErr: errors.New("boom")},
	}

	a := &OvsAdapter{ovsPath: ovsVsctlPath, ofctlPath: ovsOfctlPath, logger: slog.Default(), exec: fx}

	err := a.CreateBridge(context.Background(), "br0", 0)
	require.Error(t, err)
	require.True(t, apperrors.Is(err, apperrors.Internal))
}

func TestOvsAdapterListBridgesEmptyOutput(t *testing.T) {
	fx := &fakeExecer{
		cmd: &fakeCmd{out: []byte("\n")},
	}

	a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}
	bridges, err := a.ListBridges(context.Background())
	require.NoError(t, err)
	require.Len(t, bridges, 0)
}

func TestOvsAdapterAddFlowRuleInvalidBridge(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ofctlPath: ovsOfctlPath, logger: slog.Default(), exec: fx}

	err := a.AddFlowRule(context.Background(), badBridge, ports.FlowRule{Priority: 1, Match: "ip", Actions: "drop"})
	require.Error(t, err)
	require.True(t, apperrors.Is(err, apperrors.InvalidInput))
}

func TestOvsAdapterAddFlowRuleSuccess(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ofctlPath: ovsOfctlPath, logger: slog.Default(), exec: fx}

	err := a.AddFlowRule(context.Background(), "br0", ports.FlowRule{Priority: 100, Match: "ip", Actions: "normal"})
	require.NoError(t, err)
}

func TestOvsAdapterAddPort(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

	err := a.AddPort(context.Background(), "br0", "port1")
	require.NoError(t, err)
}

func TestOvsAdapterDeletePort(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

	err := a.DeletePort(context.Background(), "br0", "port1")
	require.NoError(t, err)
}

func TestOvsAdapterPing(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

	err := a.Ping(context.Background())
	require.NoError(t, err)
}

func TestOvsAdapterType(t *testing.T) {
	a := &OvsAdapter{}
	require.Equal(t, "ovs", a.Type())
}

func TestOvsAdapterDeleteBridge(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

	err := a.DeleteBridge(context.Background(), "br0")
	require.NoError(t, err)
}

func TestOvsAdapterDeleteFlowRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := &fakeExecer{cmd: &fakeCmd{}}
		a := &OvsAdapter{ofctlPath: ovsOfctlPath, logger: slog.Default(), exec: fx}

		err := a.DeleteFlowRule(context.Background(), "br0", "match")
		require.NoError(t, err)
	})

	t.Run("invalid bridge", func(t *testing.T) {
		fx := &fakeExecer{cmd: &fakeCmd{}}
		a := &OvsAdapter{ofctlPath: ovsOfctlPath, logger: slog.Default(), exec: fx}

		err := a.DeleteFlowRule(context.Background(), badBridge, "match")
		require.Error(t, err)
		require.True(t, apperrors.Is(err, apperrors.InvalidInput))
	})
}

func TestOvsAdapterCreateVXLANTunnel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fx := &fakeExecer{cmd: &fakeCmd{}}
		a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

		err := a.CreateVXLANTunnel(context.Background(), "br0", 100, testutil.TestVXLANRemoteIP)
		require.NoError(t, err)
	})

	t.Run("invalid bridge", func(t *testing.T) {
		fx := &fakeExecer{cmd: &fakeCmd{}}
		a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

		err := a.CreateVXLANTunnel(context.Background(), badBridge, 100, testutil.TestVXLANRemoteIP)
		require.Error(t, err)
		require.True(t, apperrors.Is(err, apperrors.InvalidInput))
	})
}

func TestOvsAdapterDeleteVXLANTunnel(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

	err := a.DeleteVXLANTunnel(context.Background(), "br0", testutil.TestVXLANRemoteIP)
	require.NoError(t, err)
}

func TestOvsAdapterCreateVethPair(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{logger: slog.Default(), exec: fx}

	err := a.CreateVethPair(context.Background(), "veth0", "veth1")
	require.NoError(t, err)
}

func TestOvsAdapterAttachVethToBridge(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{ovsPath: ovsVsctlPath, logger: slog.Default(), exec: fx}

	err := a.AttachVethToBridge(context.Background(), "br0", "veth0")
	require.NoError(t, err)
}

func TestOvsAdapterDeleteVethPair(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{logger: slog.Default(), exec: fx}

	err := a.DeleteVethPair(context.Background(), "veth0")
	require.NoError(t, err)
}

func TestOvsAdapterSetVethIP(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{}}
	a := &OvsAdapter{logger: slog.Default(), exec: fx}

	err := a.SetVethIP(context.Background(), "veth0", testutil.TestIPHost, "24")
	require.NoError(t, err)
}

func TestOvsAdapterListFlowRules(t *testing.T) {
	fx := &fakeExecer{cmd: &fakeCmd{out: []byte("cookie=0x0, duration=1.0s, table=0, n_packets=0, n_bytes=0, priority=100,ip actions=NORMAL\n")}}
	a := &OvsAdapter{ofctlPath: ovsOfctlPath, logger: slog.Default(), exec: fx}

	rules, err := a.ListFlowRules(context.Background(), "br0")
	require.NoError(t, err)
	require.NotNil(t, rules)
}
