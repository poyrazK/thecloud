// Package httputil provides HTTP helper utilities.
package httputil

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
)

// RequireTenant ensures a tenant context is set for all requests.
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantID(c)
		if tenantID == uuid.Nil {
			Error(c, errors.New(errors.InvalidInput, "X-Tenant-ID header required or default tenant not set"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// TenantMember validates that the current user is a member of the active tenant.
func TenantMember(tenantSvc ports.TenantService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		tenantID := GetTenantID(c)

		if userID == uuid.Nil || tenantID == uuid.Nil {
			Error(c, errors.New(errors.Unauthorized, "authentication and tenant required"))
			c.Abort()
			return
		}

		membership, err := tenantSvc.GetMembership(c.Request.Context(), tenantID, userID)
		if err != nil || membership == nil {
			Error(c, errors.New(errors.Forbidden, "not a member of this tenant"))
			c.Abort()
			return
		}

		c.Set("tenantRole", membership.Role)
		c.Next()
	}
}
