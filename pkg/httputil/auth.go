// Package httputil provides HTTP utilities and middleware.
package httputil

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// Auth enforces API key authentication and injects user context.
func Auth(svc ports.IdentityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			Error(c, errors.New(errors.Unauthorized, "API key required"))
			c.Abort()
			return
		}

		apiKeyObj, err := svc.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			Error(c, errors.New(errors.Unauthorized, "invalid API key"))
			c.Abort()
			return
		}

		// Wrap the request context with UserID
		ctx := appcontext.WithUserID(c.Request.Context(), apiKeyObj.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("userID", apiKeyObj.UserID) // Also keep in Gin context for convenience
		c.Next()
	}
}

// GetUserID returns the authenticated user ID from the request context.
func GetUserID(c *gin.Context) uuid.UUID {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil
	}
	userID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}

// Permission enforces RBAC checks for the provided permission.
func Permission(rbac ports.RBACService, permission domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			Error(c, errors.New(errors.Unauthorized, "authentication required"))
			c.Abort()
			return
		}

		if err := rbac.Authorize(c.Request.Context(), userID, permission); err != nil {
			Error(c, err)
			c.Abort()
			return
		}

		c.Next()
	}
}
