package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

// ctxKey represents the type of value for the context key.
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
type Handler func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct
type App struct {
	*httptreemux.TreeMux
	shutdown chan os.Signal
	log      *log.Logger
	mw       []Middleware
}

// New creates an App value that handle a set of routes for the application.
func New(shutdown chan os.Signal, log *log.Logger, mw ...Middleware) *App {
	return &App{
		TreeMux:  httptreemux.New(),
		shutdown: shutdown,
		log:      log,
		mw:       mw,
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.log.Println("error returned from handler indicated integrity issue, shutting down service")
	a.shutdown <- syscall.SIGSTOP
}

// Handle is our mechanism for mounting Handlers for a given HTTP verb and path
// pair, this makes for really easy, convenient routing.
func (a *App) Handle(verb, path string, handler Handler, mw ...Middleware) {

	// Wrap up the application-wide first, this will call the first function
	// of each middleware which will return a function of type Handler.
	handler = wrapMiddleware(wrapMiddleware(handler, mw), a.mw)

	// The function to execute for each request.
	h := func(w http.ResponseWriter, r *http.Request, params map[string]string) {

		// This API is using pointer semantic methods on this empty
		// struct type :( This is causing the need to declare this
		// variable here at the top.
		var hf tracecontext.HTTPFormat

		// Check the request for an existing Trace. The WithSpanContext
		// function can unmarshal any existing context or create a new one.
		var ctx context.Context
		var span *trace.Span
		if sc, ok := hf.SpanContextFromRequest(r); ok {
			ctx, span = trace.StartSpanWithRemoteParent(r.Context(), "internal.platform.web", sc)
		} else {
			ctx, span = trace.StartSpan(r.Context(), "internal.platform.web")
		}

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
		hf.SpanContextToRequest(span.SpanContext(), r)

		// Call the wrapped handler functions.
		if err := handler(ctx, a.log, w, r, params); err != nil {
			a.log.Printf("*****> critical shutdown error: %v", err)
			a.SignalShutdown()
			return
		}
	}

	// Add this handler for the specified verb and route.
	a.TreeMux.Handle(verb, path, h)
}

// shutdown is a type used to help with the graceful termination of the service.
type shutdown struct {
	Message string
}

// Error is the implementation of the error interface.
func (s *shutdown) Error() string {
	return s.Message
}

// Shutdown returns an error that causes the framework to signal
// a graceful shutdown.
func Shutdown(message string) error {
	return &shutdown{message}
}

// IsShutdown checks to see if the shutdown error is contained
// in the specified error value.
func IsShutdown(err error) bool {
	if _, ok := err.(*shutdown); ok {
		return true
	}
	return false
}
