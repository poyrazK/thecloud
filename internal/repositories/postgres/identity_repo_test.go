//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityRepository_Integration(t *testing.T) {
	db, _ := SetupDB(t)
	defer db.Close()
	repo := NewIdentityRepository(db)
	ctx := SetupTestUser(t, db)
	userID := appcontext.UserIDFromContext(ctx)
	tenantID := appcontext.TenantIDFromContext(ctx)

	// Cleanup
	_, _ = db.Exec(context.Background(), "DELETE FROM api_keys")


	var keyID uuid.UUID
	keyString := "test-api-key-12345"
	keyHash := sha256.Sum256([]byte(keyString))
	keyHashHex := hex.EncodeToString(keyHash[:])

	t.Run("CreateAPIKey", func(t *testing.T) {
		keyID = uuid.New()
		apiKey := &domain.APIKey{
			ID:        keyID,
			UserID:    userID, TenantID:  tenantID,
			Key:       keyString,
			KeyHash:   keyHashHex,
			Name:      "test-key",
			CreatedAt: time.Now(),
		}

		err := repo.CreateAPIKey(ctx, apiKey)
		require.NoError(t, err)
	})

	t.Run("GetAPIKeyByHash", func(t *testing.T) {
		apiKey, err := repo.GetAPIKeyByHash(ctx, keyHashHex)
		require.NoError(t, err)
		assert.Equal(t, keyID, apiKey.ID)
		assert.Equal(t, userID, apiKey.UserID)
		assert.Equal(t, "test-key", apiKey.Name)
	})

	t.Run("GetAPIKeyByID", func(t *testing.T) {
		apiKey, err := repo.GetAPIKeyByID(ctx, keyID)
		require.NoError(t, err)
		assert.Equal(t, keyString, apiKey.Key)
		assert.Equal(t, "test-key", apiKey.Name)
	})

	t.Run("ListAPIKeysByUserID", func(t *testing.T) {
		// Create another API key with dynamically generated values
		anotherKey := "test-key-" + uuid.New().String()
		anotherHash := sha256.Sum256([]byte(anotherKey))
		key2 := &domain.APIKey{
			ID:        uuid.New(),
			UserID:    userID, TenantID:  tenantID,
			Key:       anotherKey,
			KeyHash:   hex.EncodeToString(anotherHash[:]),
			Name:      "another-key",
			CreatedAt: time.Now(),
		}
		err := repo.CreateAPIKey(ctx, key2)
		require.NoError(t, err)

		keys, err := repo.ListAPIKeysByUserID(ctx, userID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(keys), 2)

		// Verify we can find our keys
		foundTestKey := false
		foundAnotherKey := false
		for _, k := range keys {
			if k.Name == "test-key" {
				foundTestKey = true
			}
			if k.Name == "another-key" {
				foundAnotherKey = true
			}
		}
		assert.True(t, foundTestKey)
		assert.True(t, foundAnotherKey)
	})

	t.Run("DeleteAPIKey", func(t *testing.T) {
		err := repo.DeleteAPIKey(ctx, keyID)
		require.NoError(t, err)

		_, err = repo.GetAPIKeyByID(ctx, keyID)
		assert.Error(t, err)
	})

	t.Run("GetAPIKeyByHash_NotFound", func(t *testing.T) {
		_, err := repo.GetAPIKeyByHash(ctx, "non-existent-hash")
		assert.Error(t, err)
	})

	t.Run("GetAPIKeyByID_NotFound", func(t *testing.T) {
		_, err := repo.GetAPIKeyByID(ctx, uuid.New())
		assert.Error(t, err)
	})
}
