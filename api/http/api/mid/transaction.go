package mid

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/api/transaction"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// BeginCommitRollback executes the transaction middleware functionality.
func BeginCommitRollback(log *logger.Logger, bgn transaction.Beginner) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			hdl := func(ctx context.Context) error {
				if err := handler(ctx, w, r); err != nil {
					return fmt.Errorf("EXECUTE TRANSACTION: %w", err)
				}
				return nil
			}

			return mid.BeginCommitRollback(ctx, log, bgn, hdl)
		}

		return h
	}

	return m
}
