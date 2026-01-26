// Package appcontext provides context utilities for the application.
package appcontext

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	tenantIDKey contextKey = "tenant_id"
)

// WithUserID returns a new context with the given userID.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext returns the userID from the context, or uuid.Nil if not found.
func UserIDFromContext(ctx context.Context) uuid.UUID {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}

// WithTenantID returns a new context with the given tenantID.
func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// TenantIDFromContext returns the tenantID from the context, or uuid.Nil if not found.
func TenantIDFromContext(ctx context.Context) uuid.UUID {
	tenantID, ok := ctx.Value(tenantIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return tenantID
}
