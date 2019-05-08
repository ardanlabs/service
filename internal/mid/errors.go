package mid

import (
	"context"
	"log"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *log.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(before web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Errors")
			defer span.End()

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.Shutdown("web value missing from context")
			}

			if err := before(ctx, w, r, params); err != nil {

				// If the error was of the type *StatusError, the handler has
				// a specific status code and error to return. If not, the
				// handler send any arbitrary error value so use 500.
				webErr, ok := errors.Cause(err).(*web.Error)
				if !ok {
					webErr = &web.Error{
						Err:    err,
						Status: http.StatusInternalServerError,
						Fields: nil,
					}
				}

				// Log the error.
				log.Printf("%s : ERROR : %+v", v.TraceID, err)

				// String provides "human readable" error messages that are
				// intended for service users to see. If the status code is
				// 500 or higher (the default) then use a generic error message.
				var errStr string
				if webErr.Status < http.StatusInternalServerError {
					errStr = webErr.Err.Error()
				} else {
					errStr = http.StatusText(webErr.Status)
				}

				// Respond with the error type we send to clients.
				res := web.ErrorResponse{
					Error:  errStr,
					Fields: webErr.Fields,
				}
				if err := web.Respond(ctx, w, res, webErr.Status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shutdown the service.
				if ok := web.IsShutdown(err); ok {
					return err
				}
			}

			// The error has been handled so we can stop propagating it.
			return nil
		}

		return h
	}

	return f
}
