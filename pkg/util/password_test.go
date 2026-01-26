package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomPassword(t *testing.T) {
	length := 16
	password, err := GenerateRandomPassword(length)
	assert.NoError(t, err)
	assert.Len(t, password, length)
	assert.NotEmpty(t, password)

	for _, ch := range password {
		assert.True(t, strings.ContainsRune(passwordChars, ch), "password contains invalid character")
	}
}

func TestGenerateRandomPasswordZeroLength(t *testing.T) {
	password, err := GenerateRandomPassword(0)
	assert.NoError(t, err)
	assert.Equal(t, "", password)
}
