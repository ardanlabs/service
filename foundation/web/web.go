// Package web contains a small web framework extension.
package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	mux    *http.ServeMux
	otmux  http.Handler
	mw     []MidHandler
	tracer trace.Tracer
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(tracer trace.Tracer, mw ...MidHandler) *App {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/trusted.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/

	mux := http.NewServeMux()

	return &App{
		mux:    mux,
		otmux:  otelhttp.NewHandler(mux, "request"),
		mw:     mw,
		tracer: tracer,
	}
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
func (a *App) EnableCORS(mw MidHandler) {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return Respond(ctx, w, "OK", http.StatusOK)
	}
	handler = wrapMiddleware([]MidHandler{mw}, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		handler(r.Context(), w, r)
	}

	finalPath := fmt.Sprintf("%s %s", http.MethodOptions, "/")

	a.mux.HandleFunc(finalPath, h)
}

// HandleNoMiddleware sets a handler function for a given HTTP method and path pair
// to the application server mux. Does not include the application middleware or
// OTEL tracing.
func (a *App) HandleNoMiddleware(method string, group string, path string, handler Handler) {
	h := func(w http.ResponseWriter, r *http.Request) {
		v := Values{
			TraceID: uuid.NewString(),
			Tracer:  nil,
			Now:     time.Now().UTC(),
		}
		ctx := setValues(r.Context(), &v)

		if err := handler(ctx, w, r); err != nil {

			// The http package will capture this panic and we will see
			// this in the logs. If the app is handling errors, this
			// should never happen.
			panic(err)
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, h)
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...MidHandler) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.startSpan(w, r)
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Tracer:  a.tracer,
			Now:     time.Now().UTC(),
		}
		ctx = setValues(ctx, &v)

		if err := handler(ctx, w, r); err != nil {

			// The http package will capture this panic and we will see
			// this in the logs. If the app is handling errors, this
			// should never happen.
			panic(err)
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, h)
}

// startSpan initializes the request by adding a span and writing otel
// related information into the response writer for the trusted.
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

	// Inject the trace information into the trusted.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

	return ctx, span
}
