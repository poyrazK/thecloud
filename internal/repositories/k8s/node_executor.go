package k8s

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"golang.org/x/crypto/ssh"
)

// NodeExecutor defines an interface for executing commands on a cluster node.
type NodeExecutor interface {
	Run(ctx context.Context, cmd string) (string, error)
	WaitForReady(ctx context.Context, timeout time.Duration) error
}

// ServiceExecutor uses the InstanceService.Exec (for Docker backend).
type ServiceExecutor struct {
	svc    ports.InstanceService
	instID uuid.UUID
}

func NewServiceExecutor(svc ports.InstanceService, instID uuid.UUID) *ServiceExecutor {
	return &ServiceExecutor{svc: svc, instID: instID}
}

func (e *ServiceExecutor) Run(ctx context.Context, cmd string) (string, error) {
	return e.svc.Exec(ctx, e.instID.String(), []string{"sh", "-c", cmd})
}

func (e *ServiceExecutor) WaitForReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for node to be ready: %w", ctx.Err())
		case <-ticker.C:
			_, err := e.Run(ctx, "echo ready")
			if err == nil {
				return nil
			}
		}
	}
}

// SSHExecutor uses SSH to run commands on a node.
type SSHExecutor struct {
	ip   string
	user string
	key  string
}

func NewSSHExecutor(ip, user, key string) *SSHExecutor {
	return &SSHExecutor{ip: ip, user: user, key: key}
}

func (e *SSHExecutor) Run(ctx context.Context, cmd string) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(e.key))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: e.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// Ephemeral nodes have dynamic keys, so strict checking isn't feasible here without a central CA.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(e.ip, "22")
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("failed to dial ssh: %w", err)
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer func() { _ = session.Close() }()

	var stdout, stderr strings.Builder
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		return stdout.String() + stderr.String(), fmt.Errorf("command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

func (e *SSHExecutor) WaitForReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for node to be ready: %w", ctx.Err())
		case <-ticker.C:
			_, err := e.Run(ctx, "echo ready")
			if err == nil {
				return nil
			}
		}
	}
}
