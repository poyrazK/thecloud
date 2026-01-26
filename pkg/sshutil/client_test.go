package sshutil

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestKey generates a private key for testing
func generateTestKey(t *testing.T) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return string(keyPEM)
}

const (
	testLocalhostSSH = "localhost:22"
	testLoopbackAddr = "127.0.0.1:0"
)

func TestNewClientWithKey(t *testing.T) {
	// 1. Valid Key
	privKey := generateTestKey(t)
	client, err := NewClientWithKey(testLocalhostSSH, "user", privKey)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, testLocalhostSSH, client.Host)
	assert.Equal(t, "user", client.User)
	assert.NotEmpty(t, client.Auth)

	// 2. Invalid Key
	_, err = NewClientWithKey(testLocalhostSSH, "user", "invalid-key")
	require.Error(t, err)
}

func TestWaitForSSH(t *testing.T) {
	// Start a dummy TCP server
	l, err := net.Listen("tcp", testLoopbackAddr)
	require.NoError(t, err)
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port
	host := fmt.Sprintf("127.0.0.1:%d", port)

	client := &Client{Host: host}

	// Should connect successfully
	err = client.WaitForSSH(context.Background(), 2*time.Second)
	require.NoError(t, err)
}

func TestWaitForSSHTimeout(t *testing.T) {
	// Pick a random port (hopefully unused)
	client := &Client{Host: "127.0.0.1:54321"} // Unlikely to be a valid SSH server immediately

	// Should timeout
	// Use small timeout for test speed
	err := client.WaitForSSH(context.Background(), 100*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestRunConnectionRefused(t *testing.T) {
	// Ensure we pick a port that rejects connection
	l, err := net.Listen("tcp", testLoopbackAddr)
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close() // Close immediately to ensure connection refused

	host := fmt.Sprintf("127.0.0.1:%d", port)
	privKey := generateTestKey(t)
	client, _ := NewClientWithKey(host, "user", privKey)

	_, err = client.Run(context.Background(), "echo hello")
	require.Error(t, err)
	// Error message format depends on OS, usually "connection refused" or "dial tcp"
}

// TODO: A full SSH server mock for Run and WriteFile would be better but significantly more complex.
// For "Phase 1 Quick Wins", validating the Client logic, Key parsing, and Network dialing is a good start.

func TestWriteFileConnectionRefused(t *testing.T) {
	// Pick a random port (hopefully unused)
	client := &Client{Host: testLoopbackAddr}

	err := client.WriteFile(context.Background(), "/tmp/test", []byte("data"), "0644")
	require.Error(t, err)
	// Expect dial error
}
