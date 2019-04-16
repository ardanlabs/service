package mid

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/ardanlabs/service/internal/platform/web"
	"go.opencensus.io/trace"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func (mw *Middleware) Errors(before web.Handler) web.Handler {

	// Create the handler that will be attached in the middleware chain.
	h := func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		ctx, span := trace.StartSpan(ctx, "internal.mid.ErrorHandler")
		defer span.End()

		// If the context is missing this value, request the service
		// to be shutdown gracefully.
		v, ok := ctx.Value(web.KeyValues).(*web.Values)
		if !ok {
			return web.Shutdown("web value missing from context")
		}

		// In the event of a panic, we want to capture it here so we can send an
		// error down the stack.
		defer func() {
			if r := recover(); r != nil {

				// Indicate this request had an error.
				v.Error = true

				// Log the panic.
				log.Printf("%s : ERROR : Panic Caught : %s\n", v.TraceID, r)
				log.Printf("%s : ERROR : Stacktrace\n%s\n", v.TraceID, debug.Stack())

				// Respond with the error.
				res := web.ErrorResponse{
					Error: "unhandled error",
				}

				if err := web.Respond(ctx, log, w, res, http.StatusInternalServerError); err != nil {
					// TODO what if this fails?
				}
			}
		}()

		if err := before(ctx, log, w, r, params); err != nil {

			// Indicate this request had an error.
			v.Error = true

			// Convert the error interface variable to the concrete type
			// *web.StatusError to find the appropriate HTTP status.
			statusError := web.NewStatusError(err)

			// If the error is an internal issue then log the error message.
			// Do not log error messages that come from client requests.
			if statusError.Status >= http.StatusInternalServerError {
				log.Printf("%s : %+v", v.TraceID, err)
			}

			// Respond with the error type we send to clients.
			res := web.ErrorResponse{
				Error:  statusError.String(),
				Fields: statusError.Fields,
			}

			if err := web.Respond(ctx, log, w, res, statusError.Status); err != nil {
				return err
			}

			// If we receive the shutdown err we need to return it
			// back to the base handler to shutdown the service.
			if ok := web.IsShutdown(err); ok {
				return err
			}

			// The error has been handled so we can stop propagating it.
			return nil
		}

		return nil
	}

	return h
}
