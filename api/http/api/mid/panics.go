package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/web"
)

// Panics executes the panic middleware functionality.
func Panics() web.MidHandler {
	midFunc := func(ctx context.Context, w http.ResponseWriter, r *http.Request, hdl mid.Handler) (any, error) {
		return mid.Panics(ctx, hdl)
	}

	return addMiddleware(midFunc)
}
