// Package k8s implements Kubernetes provisioning adapters.
package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/pkg/sshutil"
)

// NodeExecutor abstracts the execution of commands on a node.
type NodeExecutor interface {
	Run(ctx context.Context, cmd string) (string, error)
	WaitForReady(ctx context.Context, timeout time.Duration) error
}

// SSHExecutor implements NodeExecutor using SSH.
type SSHExecutor struct {
	ip   string
	user string
	key  string
}

// NewSSHExecutor creates a new SSHExecutor.
func NewSSHExecutor(ip, user, key string) *SSHExecutor {
	return &SSHExecutor{ip: ip, user: user, key: key}
}

const sshRetryInterval = 2 * time.Second

func (e *SSHExecutor) Run(ctx context.Context, cmd string) (string, error) {
	client, err := sshutil.NewClientWithKey(e.ip, e.user, e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create ssh client: %w", err)
	}
	return client.Run(ctx, cmd) // Propagate ctx
}

func (e *SSHExecutor) WaitForReady(ctx context.Context, timeout time.Duration) error {
	client, err := sshutil.NewClientWithKey(e.ip, e.user, e.key)
	if err != nil {
		return fmt.Errorf("failed to create ssh client: %w", err)
	}
	// We pass the original ctx here because WaitForSSH now handles its own timeout context internally if needed,
	// but actually our updated signature expects (ctx, timeout).
	// The original code created a separate timeout context but didn't pass it.
	// Now we pass the parent context and let WaitForSSH handle the timeout loop with the duration.
	return client.WaitForSSH(ctx, timeout)
}

// ServiceExecutor implements NodeExecutor using InstanceService.Exec.
type ServiceExecutor struct {
	instSvc ports.InstanceService
	instID  uuid.UUID
}

// NewServiceExecutor creates a new ServiceExecutor.
func NewServiceExecutor(instSvc ports.InstanceService, instID uuid.UUID) *ServiceExecutor {
	return &ServiceExecutor{instSvc: instSvc, instID: instID}
}

func (e *ServiceExecutor) Run(ctx context.Context, cmd string) (string, error) {
	// InstanceService.Exec expects []string for arguments.
	// Since we are passing a shell command string, we need to wrap it in /bin/sh -c "cmd"
	// because ComputeBackend.Exec normally executes a binary with args directly.
	// Docker exec: ["/bin/sh", "-c", cmd]
	wrappedCmd := []string{"/bin/sh", "-c", cmd}
	return e.instSvc.Exec(ctx, e.instID.String(), wrappedCmd)
}

func (e *ServiceExecutor) WaitForReady(ctx context.Context, timeout time.Duration) error {
	// For Docker, if the container is running, it's "ready" enough for Exec usually.
	// We can try a simple echo.
	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Check immediately once
	if _, err := e.Run(ctx, "echo ready"); err == nil {
		return nil
	}

	ticker := time.NewTicker(sshRetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctxTimeout.Done():
			return fmt.Errorf("timeout waiting for node ready")
		case <-ticker.C:
			_, err := e.Run(ctx, "echo ready")
			if err == nil {
				return nil
			}
		}
	}
}
