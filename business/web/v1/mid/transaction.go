package mid

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/sys/database"
	"github.com/ardanlabs/service/business/sys/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// ExecuteInTransation starts a transaction around all the storage calls within
// the scope of the handler function.
func ExecuteInTransation(log *logger.Logger, bgn database.Beginner) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			log.Info(ctx, "BEGIN TRANSACTION")
			tx, err := bgn.Begin()
			if err != nil {
				return err
			}

			defer func() {
				log.Info(ctx, "CHECKING FOR ROLLBACK")
				if err := tx.Rollback(); err != nil {
					if errors.Is(err, sql.ErrTxDone) {
						return
					}
					log.Info(ctx, "ROLLBACK TRANSACTION", "ERROR", err)
				}
			}()

			ctx = database.SetTransaction(ctx, tx)

			if err := handler(ctx, w, r); err != nil {
				return err
			}

			log.Info(ctx, "COMMIT TRANSACTION")
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit tran: %w", err)
			}

			return nil
		}

		return h
	}

	return m
}
