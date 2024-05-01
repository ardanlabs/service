package mid

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ardanlabs/service/business/api/transaction"
	"github.com/ardanlabs/service/foundation/logger"
)

// BeginCommitRollback starts a transaction around all the calls within the
// scope of the handler function.
func BeginCommitRollback(ctx context.Context, log *logger.Logger, bgn transaction.Beginner, handler Handler) error {
	hasCommitted := false

	log.Info(ctx, "BEGIN TRANSACTION")
	tx, err := bgn.Begin()
	if err != nil {
		return fmt.Errorf("BEGIN TRANSACTION: %w", err)
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

	ctx = setTran(ctx, tx)

	if err := handler(ctx); err != nil {
		return fmt.Errorf("EXECUTE TRANSACTION: %w", err)
	}

	log.Info(ctx, "COMMIT TRANSACTION")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("COMMIT TRANSACTION: %w", err)
	}

	hasCommitted = true

	return nil
}
