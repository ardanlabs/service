// Package web contains a small web framework extension.
package web

import (
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// A Handler is a type that handles a http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	mux      *httptreemux.ContextMux
	otmux    http.Handler
	shutdown chan os.Signal
	mw       []Middleware
	tracer   trace.Tracer
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, tracer trace.Tracer, mw ...Middleware) *App {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/response.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/

	mux := httptreemux.NewContextMux()

	return &App{
		mux:      mux,
		otmux:    otelhttp.NewHandler(mux, "request"),
		shutdown: shutdown,
		mw:       mw,
		tracer:   tracer,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This was set up on line 44 in the NewApp function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.otmux.ServeHTTP(w, r)
}

// EnableCORS enables CORS preflight requests to work in the middleware. It
// prevents the MethodNotAllowedHandler from being called. This must be enabled
// for the CORS middleware to work.
func (a *App) EnableCORS(mw Middleware) {
	a.mw = append(a.mw, mw)

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return Respond(ctx, w, "OK", http.StatusOK)
	}
	handler = wrapMiddleware(a.mw, handler)

	a.mux.OptionsHandler = func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		ctx, span := a.startSpan(w, r)
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Tracer:  a.tracer,
			Now:     time.Now().UTC(),
		}
		ctx = SetValues(ctx, &v)

		handler(ctx, w, r)
	}
}

// HandleNoMiddleware sets a handler function for a given HTTP method and path pair
// to the application server mux. Does not include the application middleware.
func (a *App) HandleNoMiddleware(method string, group string, path string, handler Handler) {
	a.handle(method, group, path, handler)
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	a.handle(method, group, path, handler)
}

// =============================================================================

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) handle(method string, group string, path string, handler Handler) {
	h := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.startSpan(w, r)
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Tracer:  a.tracer,
			Now:     time.Now().UTC(),
		}
		ctx = SetValues(ctx, &v)

		if err := handler(ctx, w, r); err != nil {
			if validateShutdown(err) {
				a.SignalShutdown()
				return
			}
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.mux.Handle(method, finalPath, h)
}

// startSpan initializes the request by adding a span and writing otel
// related information into the response writer for the response.
func (a *App) startSpan(w http.ResponseWriter, r *http.Request) (context.Context, trace.Span) {
	ctx := r.Context()

	// There are times when the handler is called without a tracer, such
	// as with tests. We need a span for the trace id.
	span := trace.SpanFromContext(ctx)

	// If a tracer exists, then replace the span for the one currently
	// found in the context. This may have come from over the wire.
	if a.tracer != nil {
		ctx, span = a.tracer.Start(ctx, "pkg.web.handle")
		span.SetAttributes(attribute.String("endpoint", r.RequestURI))
	}

	// Inject the trace information into the response.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

	return ctx, span
}

// validateShutdown validates the error for special conditions that do not
// warrant an actual shutdown by the system.
func validateShutdown(err error) bool {

	// Ignore syscall.EPIPE and syscall.ECONNRESET errors which occurs
	// when a write operation happens on the http.ResponseWriter that
	// has simultaneously been disconnected by the client (TCP
	// connections is broken). For instance, when large amounts of
	// data is being written or streamed to the client.
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// https://gosamples.dev/broken-pipe/
	// https://gosamples.dev/connection-reset-by-peer/

	switch {
	case errors.Is(err, syscall.EPIPE):

		// Usually, you get the broken pipe error when you write to the connection after the
		// RST (TCP RST Flag) is sent.
		// The broken pipe is a TCP/IP error occurring when you write to a stream where the
		// other end (the peer) has closed the underlying connection. The first write to the
		// closed connection causes the peer to reply with an RST packet indicating that the
		// connection should be terminated immediately. The second write to the socket that
		// has already received the RST causes the broken pipe error.
		return false

	case errors.Is(err, syscall.ECONNRESET):

		// Usually, you get connection reset by peer error when you read from the
		// connection after the RST (TCP RST Flag) is sent.
		// The connection reset by peer is a TCP/IP error that occurs when the other end (peer)
		// has unexpectedly closed the connection. It happens when you send a packet from your
		// end, but the other end crashes and forcibly closes the connection with the RST
		// packet instead of the TCP FIN, which is used to close a connection under normal
		// circumstances.
		return false
	}

	return true
}
