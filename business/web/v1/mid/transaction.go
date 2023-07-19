package mid

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/data/transaction"
	"github.com/ardanlabs/service/business/sys/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// ExecuteInTransation starts a transaction around all the storage calls within
// the scope of the handler function.
func ExecuteInTransation(log *logger.Logger, bgn transaction.Beginner) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			hasCommited := false

			log.Info(ctx, "BEGIN TRANSACTION")
			tx, err := bgn.Begin()
			if err != nil {
				return fmt.Errorf("BEGIN TRANSACTION: %w", err)
			}

			defer func() {
				if !hasCommited {
					log.Info(ctx, "ROLLBACK TRANSACTION")
				}

				if err := tx.Rollback(); err != nil {
					if errors.Is(err, sql.ErrTxDone) {
						return
					}
					log.Info(ctx, "ROLLBACK TRANSACTION", "ERROR", err)
				}
			}()

			ctx = transaction.Set(ctx, tx)

			if err := handler(ctx, w, r); err != nil {
				return fmt.Errorf("EXECUTE TRANSACTION: %w", err)
			}

			log.Info(ctx, "COMMIT TRANSACTION")
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("COMMIT TRANSACTION: %w", err)
			}

			hasCommited = true

			return nil
		}

		return h
	}

	return m
}
