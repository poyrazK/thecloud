package sshutil

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client for remote execution.
type Client struct {
	Host            string
	User            string
	Auth            []ssh.AuthMethod
	HostKeyCallback ssh.HostKeyCallback
}

// NewClientWithKey constructs an SSH client using a private key.
func NewClientWithKey(host, user, privateKey string) (*Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Client{
		Host: host,
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Default to insecure for legacy compatibility, but now configurable
	}, nil
}

// Run executes a command and returns its output.
func (c *Client) Run(ctx context.Context, cmd string) (string, error) {
	config := &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		HostKeyCallback: c.HostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := c.Host
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "22")
	}

	d := net.Dialer{Timeout: config.Timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to dial: %w", err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return "", fmt.Errorf("failed to create ssh client conn: %w", err)
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b

	// We can't easily cancel session.Run directly with context, but the underlying connection closure will abort it.
	// Alternatively, we could start the command and wait for it or context.
	if err := session.Run(cmd); err != nil {
		return b.String(), fmt.Errorf("failed to run command %q: %w", cmd, err)
	}

	return b.String(), nil
}

// WriteFile writes content to a remote file.
func (c *Client) WriteFile(ctx context.Context, path string, content []byte, mode string) error {
	config := &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		HostKeyCallback: c.HostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := c.Host
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "22")
	}

	d := net.Dialer{Timeout: config.Timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create ssh client conn: %w", err)
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		w, err := session.StdinPipe()
		if err != nil {
			errCh <- fmt.Errorf("failed to get stdin pipe: %w", err)
			return
		}
		defer w.Close()

		filename := filepath.Base(path)
		fmt.Fprintf(w, "C%s %d %s\n", mode, len(content), filename)
		if _, err := w.Write(content); err != nil {
			errCh <- fmt.Errorf("failed to write content: %w", err)
			return
		}
		if _, err := fmt.Fprint(w, "\x00"); err != nil {
			errCh <- fmt.Errorf("failed to write null byte: %w", err)
			return
		}
	}()

	if err := session.Run("/usr/bin/scp -t " + path); err != nil {
		// Check if we had a goroutine error first
		select {
		case pipeErr := <-errCh:
			if pipeErr != nil {
				return pipeErr
			}
		default:
		}
		return fmt.Errorf("failed to scp: %w", err)
	}

	// Double check for any pipe errors after run completes
	if err := <-errCh; err != nil {
		return err
	}

	return nil
}

const (
	sshWaitInterval = 2 * time.Second
	sshDialTimeout  = 2 * time.Second
)

// WaitForSSH waits for the SSH port to be open and accepting connections.
func (c *Client) WaitForSSH(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(sshWaitInterval)
	defer ticker.Stop()

	// Determine address to dial once
	addr := c.Host
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, "22")
	}

	// Check immediately once
	d := net.Dialer{Timeout: sshDialTimeout}
	if conn, err := d.DialContext(ctx, "tcp", addr); err == nil {
		conn.Close()
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for SSH on %s: %w", c.Host, ctx.Err())
		case <-ticker.C:
			d := net.Dialer{Timeout: sshDialTimeout}
			conn, err := d.DialContext(ctx, "tcp", addr)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}
