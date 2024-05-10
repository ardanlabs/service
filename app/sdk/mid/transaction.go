package mid

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/sdk/transaction"
	"github.com/ardanlabs/service/foundation/logger"
)

// BeginCommitRollback starts a transaction for the domain call.
func BeginCommitRollback(ctx context.Context, log *logger.Logger, bgn transaction.Beginner, next Handler) (Encoder, error) {
	hasCommitted := false

	log.Info(ctx, "BEGIN TRANSACTION")
	tx, err := bgn.Begin()
	if err != nil {
		return nil, errs.Newf(errs.Internal, "BEGIN TRANSACTION: %s", err)
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

	resp, err := next(ctx)
	if err != nil {
		return nil, errs.Newf(errs.Internal, "EXECUTE TRANSACTION: %s", err)
	}

	log.Info(ctx, "COMMIT TRANSACTION")
	if err := tx.Commit(); err != nil {
		return nil, errs.Newf(errs.Internal, "COMMIT TRANSACTION: %s", err)
	}

	hasCommitted = true

	return resp, err
}
