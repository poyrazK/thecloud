//go:build e2e
// +build e2e

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAM_E2E(t *testing.T) {
	if err := waitForServer(); err != nil {
		t.Fatalf("Failing IAM E2E test: %v", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	token := registerAndLogin(t, client, "iam-tester@thecloud.local", "IAM Tester")

	var policyID string

	// 1. Create Policy
	t.Run("CreatePolicy", func(t *testing.T) {
		payload := domain.Policy{
			Name: "E2E-Test-Policy",
			Statements: []domain.Statement{
				{
					Effect:   domain.EffectDeny,
					Action:   []string{"instance:launch"},
					Resource: []string{"*"},
				},
			},
		}
		resp := postRequest(t, client, testutil.TestBaseURL+"/iam/policies", token, payload)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var res struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&res))
		policyID = res.Data.ID
		assert.NotEmpty(t, policyID)
	})

	// 2. Get Current User ID (via some profile endpoint if it exists, or just login resp)
	// For this test, we need the userID. In a real scenario, we might have a /auth/me
	// Assuming registerAndLogin gives us a token for a user we just created.
}
