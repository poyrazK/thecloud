package appcontext_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/stretchr/testify/assert"
)

func TestUserIDContext(t *testing.T) {
	t.Run("Extract from empty context", func(t *testing.T) {
		ctx := context.Background()
		userID := appcontext.UserIDFromContext(ctx)
		assert.Equal(t, uuid.Nil, userID)
	})

	t.Run("Set and extract", func(t *testing.T) {
		ctx := context.Background()
		expectedID := uuid.New()

		ctx = appcontext.WithUserID(ctx, expectedID)
		userID := appcontext.UserIDFromContext(ctx)

		assert.Equal(t, expectedID, userID)
	})
}
