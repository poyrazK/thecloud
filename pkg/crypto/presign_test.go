package crypto

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignAndVerifyURL(t *testing.T) {
	secret := "super-secret-key"
	baseURL := "http://localhost:8080"
	method := "GET"
	bucket := "mybucket"
	key := "my/file.jpg"
	expires := time.Now().Add(1 * time.Hour)

	// Test Signing
	signedURL, err := SignURL(secret, baseURL, method, bucket, key, expires)
	require.NoError(t, err)
	assert.Contains(t, signedURL, "signature=")
	assert.Contains(t, signedURL, "expires=")

	// Parse generated URL to extract params for verification
	// In real usage, the handler extracts these from the request context
	// URL format: http://localhost:8080/storage/presigned/mybucket/my/file.jpg?expires=...&signature=...

	// manually extract for test
	// path: /storage/presigned/mybucket/my/file.jpg
	path := "/storage/presigned/mybucket/my/file.jpg"

	// extract query params
	// This is a bit manual but simulates what the handler will do
	// We'll trust our SignURL implementation put the right params in

	// Let's verify using the public Verify function
	// We need to parse the URL to get the signature and expiration

	importURL, _ := time.Parse("2006-01-02", "2022-01-01") // dummy
	_ = importURL

	// 1. Valid Signature
	// Requires parsing query params.
	// Let's use net/url for parsing in test
	u, _ := url.Parse(signedURL)
	q := u.Query()
	sig := q.Get("signature")
	exp := q.Get("expires")

	err = VerifyURL(secret, method, path, exp, sig)
	require.NoError(t, err)

	// 2. Tampered Path
	err = VerifyURL(secret, method, "/other/path", exp, sig)
	require.Error(t, err)
	assert.Equal(t, "invalid signature", err.Error())

	// 3. Tampered Expiration
	err = VerifyURL(secret, method, path, "12345", sig)
	require.Error(t, err)

	// 4. Expired URL
	expiredTime := time.Now().Add(-1 * time.Hour)
	expiredURL, _ := SignURL(secret, baseURL, method, bucket, key, expiredTime)
	uExp, _ := url.Parse(expiredURL)
	qExp := uExp.Query()
	err = VerifyURL(secret, method, path, qExp.Get("expires"), qExp.Get("signature"))
	require.Error(t, err)
	assert.Equal(t, "URL expired", err.Error())

	// 5. Wrong Method (e.g. tried PUT on a GET signature)
	err = VerifyURL(secret, "PUT", path, exp, sig)
	require.Error(t, err)
}
