package web

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
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
	ClientIP   string
}

// A Handler is a type that handles an http request within our own little mini
// framework.
type Handler func(ctx context.Context, log *log.Logger, w http.ResponseWriter, r *http.Request, params map[string]string) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct
type App struct {
	*httptreemux.TreeMux
	log *log.Logger
	mw  []Middleware
}

// New creates an App value that handle a set of routes for the application.
func New(log *log.Logger, mw ...Middleware) *App {
	return &App{
		TreeMux: httptreemux.New(),
		log:     log,
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
			TraceID:  span.SpanContext().TraceID.String(),
			Now:      time.Now(),
			ClientIP: clientIPFromRequest(r),
		}
		ctx = context.WithValue(ctx, KeyValues, &v)

		// Set the parent span on the outgoing requests before any other header to
		// ensure that the trace is ALWAYS added to the request regardless of
		// any error occuring or not.
		hf.SpanContextToRequest(span.SpanContext(), r)

		// Call the wrapped handler functions.
		if err := handler(ctx, a.log, w, r, params); err != nil {
			Error(ctx, a.log, w, err)
		}
	}

	// Add this handler for the specified verb and route.
	a.TreeMux.Handle(verb, path, h)
}

// clientIPFromRequest implements effort to return the real client IP
func clientIPFromRequest(r *http.Request) string {
	// It parses X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy or AWS ALB/LB
	// Use X-Forwarded-For before X-Real-Ip as nginx uses X-Real-Ip with the proxy's IP.
	clientIP := r.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	}

	if clientIP != "" {
		return clientIP
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""

}
