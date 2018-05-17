package web

import (
	"context"
	"net/http"
	"time"

	"github.com/ardanlabs/service/internal/platform/trace"
	"github.com/dimfeld/httptreemux"
)

// TraceIDHeader is the header added to outgoing requests which adds the
// traceID to it.
const TraceIDHeader = "X-Trace-ID"

// Key represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values or stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
	Error      bool
}

// A Handler is a type that handles an http request within our own little mini
// framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct
type App struct {
	*httptreemux.TreeMux
	mw []Middleware
}

// New creates an App value that handle a set of routes for the application.
func New(mw ...Middleware) *App {
	return &App{
		TreeMux: httptreemux.New(),
		mw:      mw,
	}
}

// Handle is our mechanism for mounting Handlers for a given HTTP verb and path
// pair, this makes for really easy, convenient routing.
func (a *App) Handle(verb, path string, handler Handler, mw ...Middleware) {

	// Wrap up the application-wide first, this will call the first function
	// of each middleware which will return a function of type Handler.
	handler = wrapMiddleware(wrapMiddleware(handler, mw), a.mw)

	// The function to execute for each request.
	h := func(w http.ResponseWriter, r *http.Request, params map[string]string) {

		// Check the request for an existing Trace. The WithSpanContext
		// function can unmarshal any existing context or create a new one.
		spanContext := r.Header.Get(TraceIDHeader)
		ctx, span := trace.WithSpanContext(r.Context(), "internal.platform.web", spanContext)
		defer span.End()

		// Set the context with the required values to
		// process the request.
		v := Values{
			TraceID: span.SpanContext().TraceID.String(),
			Now:     time.Now(),
		}
		ctx = context.WithValue(ctx, KeyValues, &v)

		// Set the parent span on the outgoing requests before any other header to
		// ensure that the trace is ALWAYS added to the request regardless of
		// any error occuring or not.
		data, err := trace.MarshalSpanContext(span.SpanContext())
		if err == nil {
			w.Header().Set(TraceIDHeader, string(data))
		}

		// Call the wrapped handler functions.
		if err := handler(ctx, w, r, params); err != nil {
			Error(ctx, w, err)
		}
	}

	// Add this handler for the specified verb and route.
	a.TreeMux.Handle(verb, path, h)
}
