package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/foundation/web"
)

/*
	// Option requests must be handled for CORS. This will send a 200 Status OK
	// for all Option requests. This must be called when setting up routes in
	// sales-api/handlers/v1/v1.go
	app.OptionsHandlerAll()

	// If you want more control over the response, you can code the handler.
	app.OptionsHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})
*/

// Cors sets the response headers needed for Cross-Origin Resource Sharing
func Cors(origin string) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// Set the CORS headers to the response.
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			// Call the next handler.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
