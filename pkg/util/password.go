// Package util provides general utility functions.
package util

import (
	"crypto/rand"
	"math/big"
)

const (
	passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
)

// GenerateRandomPassword returns a random password of the requested length.
func GenerateRandomPassword(length int) (string, error) {
	password := make([]byte, length)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		if err != nil {
			return "", err
		}
		password[i] = passwordChars[num.Int64()]
	}
	return string(password), nil
}
