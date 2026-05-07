// Package core provides core business services.
package core

import (
	"context"
)

// ctxKeyTransaction is the key used to store/retrieve pgx.Tx in context.
type ctxKeyTransaction struct{}

type txKey struct{}

// WithTransaction stores a pgx.Tx in the context for transaction-aware repositories.
func WithTransaction(ctx context.Context, tx interface{}) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TransactionFromContext retrieves a pgx.Tx from the context, or nil if not present.
func TransactionFromContext(ctx context.Context) interface{} {
	return ctx.Value(txKey{})
}