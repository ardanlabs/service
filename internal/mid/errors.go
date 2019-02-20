package mid

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

// ErrorHandler for catching and responding to errors.
func ErrorHandler(before web.Handler) web.Handler {

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

				// Respond with the error.
				web.Error(ctx, log, w, errors.New("unhandled"))

				// Print out the stack.
				log.Printf("%s : ERROR : Stacktrace\n%s\n", v.TraceID, debug.Stack())
			}
		}()

		if err := before(ctx, log, w, r, params); err != nil {

			// Indicate this request had an error.
			v.Error = true

			// What is the root error.
			err = errors.Cause(err)

			// If we receive the shutdown err we need to return it
			// back to the base handler to shutdown the service.
			if ok := web.IsShutdown(err); ok {
				web.Error(ctx, log, w, errors.New("unhandled"))
				return err
			}

			// Don't log errors based on not found issues. This has
			// the potential to create noise in the logs.
			if err != web.ErrNotFound {
				log.Printf("%s : ERROR : %v\n", v.TraceID, err)
			}

			// Respond with the error.
			web.Error(ctx, log, w, err)

			// The error has been handled so we can stop propagating it.
			return nil
		}

		return nil
	}

	return h
}
