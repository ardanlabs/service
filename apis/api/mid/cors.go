package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/foundation/web"
)

// Cors sets the response headers needed for Cross-Origin Resource Sharing
func Cors(origins []string) web.MidHandler {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			for _, origin := range origins {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
