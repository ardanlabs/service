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
func BeginCommitRollback(log *logger.Logger, bgn transaction.Beginner) web.Middleware {
	midFunc := func(ctx context.Context, r *http.Request, next mid.Handler) (mid.Encoder, error) {
		return mid.BeginCommitRollback(ctx, log, bgn, next)
	}

	return addMiddleware(midFunc)
}
