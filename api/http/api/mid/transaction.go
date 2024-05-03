package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/business/api/transaction"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// BeginCommitRollback executes the transaction middleware functionality.
func BeginCommitRollback(log *logger.Logger, bgn transaction.Beginner) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			hdl := func(ctx context.Context) (any, error) {
				return handler(ctx, w, r)
			}

			return mid.BeginCommitRollback(ctx, log, bgn, hdl)
		}

		return h
	}

	return m
}
