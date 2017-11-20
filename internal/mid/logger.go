package mid

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ardanlabs/service/internal/platform/web"
)

// RequestLogger writes some information about the request to the logs in
// the format: TraceID : (200) GET /foo -> IP ADDR (latency)
func RequestLogger(next web.Handler) web.Handler {

	// Wrap this handler around the next one provided.
	h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		v := ctx.Value(web.KeyValues).(*web.Values)

		next(ctx, w, r, params)

		log.Printf("%s : (%d) : %s %s -> %s (%s)",
			v.TraceID,
			v.StatusCode,
			r.Method, r.URL.Path,
			r.RemoteAddr, time.Since(v.Now),
		)

		// This is the top of the food chain. At this point all error
		// handling has been done including logging.
		return nil
	}

	return h
}
