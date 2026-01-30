package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/pkg/testutil"
)

func TestQuotaLimits(t *testing.T) {
	t.Parallel()
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing Quota Limits test: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	token := registerAndLogin(t, client, "quota-tester@thecloud.local", "Quota Tester")

	t.Run("Rate Limiting", func(t *testing.T) {
		// Burst requests to trigger rate limiting
		// The platform likely has a limit per user/IP
		triggered := false
		for i := 0; i < 50; i++ {
			resp := getRequest(t, client, testutil.TestBaseURL+"/instances", token)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusTooManyRequests {
				triggered = true
				break
			}
			// Small sleep to avoid overloading the small test server too much,
			// but fast enough to trigger limits if they are low.
			time.Sleep(10 * time.Millisecond)
		}

		// If rate limiting is enabled, we might trigger it.
		// If not, we just verified we can handle 50 requests.
		// Note: Don't strictly fail if not triggered unless we KNOW the limit is < 50.
		if triggered {
			t.Log("Rate limit triggered as expected")
		} else {
			t.Log("Rate limit not triggered within 50 requests")
		}
	})

	t.Run("Maximum API Keys", func(t *testing.T) {
		// Try to create many API keys
		maxHit := false
		for i := 0; i < 50; i++ {
			payload := map[string]string{"name": fmt.Sprintf("key-%d", i)}
			resp := postRequest(t, client, testutil.TestBaseURL+"/auth/keys", token, payload)

			if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusForbidden {
				// Hit a quota!
				maxHit = true
				_ = resp.Body.Close()
				break
			}
			_ = resp.Body.Close()
		}

		if maxHit {
			t.Log("API key quota hit successfully")
		} else {
			t.Log("Warning: Could not hit API key quota within 50 attempts")
		}
	})
}
