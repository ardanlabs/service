// Package transaction provides support for database transaction related functionality.
package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ardanlabs/service/foundation/logger"
)

// Transaction represents a value that can commit or rollback a transaction.
type Transaction interface {
	Commit() error
	Rollback() error
}

// Beginner represents a value that can begin a transaction.
type Beginner interface {
	Begin() (Transaction, error)
}

type ctxKey int

const trKey ctxKey = 2

// Set stores a value that can manage a transaction.
func Set(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, trKey, tx)
}

// Get retrieves the value that can manage a transaction.
func Get(ctx context.Context) (Transaction, bool) {
	v, ok := ctx.Value(trKey).(Transaction)
	return v, ok
}

// ExecuteUnderTransaction is a helper function that can be used in tests and
// other apps to execute the core APIs under a transaction.
func ExecuteUnderTransaction(ctx context.Context, log *logger.Logger, bgn Beginner, fn func(tx Transaction) error) error {
	hasCommitted := false

	log.Info(ctx, "BEGIN TRANSACTION")
	tx, err := bgn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if !hasCommitted {
			log.Info(ctx, "ROLLBACK TRANSACTION")
		}

		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			log.Info(ctx, "ROLLBACK TRANSACTION", "ERROR", err)
		}
	}()

	if err := fn(tx); err != nil {
		return fmt.Errorf("EXECUTE TRANSACTION: %w", err)
	}

	log.Info(ctx, "COMMIT TRANSACTION")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("COMMIT TRANSACTION: %w", err)
	}

	hasCommitted = true

	return nil
}
