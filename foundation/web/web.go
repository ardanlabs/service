// Package web contains a small web framework extension.
package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/foundation/tracer"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// Handler represents a function that handles a http request within our own
// little mini framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error)

// Logger represents a function that will be called to add information
// to the logs.
type Logger func(ctx context.Context, msg string, v ...any)

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	log    Logger
	tracer trace.Tracer
	mux    *http.ServeMux
	otmux  http.Handler
	mw     []MidHandler
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(log Logger, tracer trace.Tracer, mw ...MidHandler) *App {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/trusted.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/

	mux := http.NewServeMux()

	return &App{
		log:    log,
		tracer: tracer,
		mux:    mux,
		otmux:  otelhttp.NewHandler(mux, "request"),
		mw:     mw,
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
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
		return struct{ Status string }{Status: "OK"}, nil
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
		}
		ctx := setValues(r.Context(), &v)

		resp, err := handler(ctx, w, r)
		if err != nil {
			send(ctx, w, err)
		}
		send(ctx, w, resp)
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
		ctx, span := tracer.StartTrace(r.Context(), a.tracer, "pkg.web.handle", r.RequestURI, w)
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
		}
		ctx = setValues(ctx, &v)

		resp, err := handler(ctx, w, r)
		if err != nil {
			send(ctx, w, err)
		}
		send(ctx, w, resp)
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	finalPath = fmt.Sprintf("%s %s", method, finalPath)

	a.mux.HandleFunc(finalPath, h)
}
