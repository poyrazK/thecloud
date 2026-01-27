package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSSHKeyPair(t *testing.T) {
	privateKey, publicKey, err := GenerateSSHKeyPair()
	require.NoError(t, err)
	assert.Contains(t, privateKey, "BEGIN RSA PRIVATE KEY")
	assert.True(t, strings.HasPrefix(strings.TrimSpace(publicKey), "ssh-rsa"))
}
