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
	midFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request, hdl mid.Handler) (any, error) {
		return mid.BeginCommitRollback(ctx, log, bgn, hdl)
	}

	return addMiddleware(midFunc)
}
