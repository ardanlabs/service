// Package core provides support for common core related functionality.
package core

import (
	"context"
)

// Transactor represents a value that can commit or rollback a transaction.
type Transactor interface {
	Commit() error
	Rollback() error
}

// Transaction represents a value that can begin a transaction.
type Transaction interface {
	Begin() (Transactor, error)
}

// =============================================================================

type ctxKey int

const trKey ctxKey = 2

// SetTransaction stores a value that can control a transaction.
func SetTransaction(ctx context.Context, tr Transactor) context.Context {
	return context.WithValue(ctx, trKey, tr)
}

// GetTransaction retrieves the value that can control a transaction.
func GetTransaction(ctx context.Context) (Transactor, bool) {
	v, ok := ctx.Value(trKey).(Transactor)
	return v, ok
}
