// Package metrics cosntructs the metrics the application will track.
package metrics

import (
	"context"
	"expvar"
	"sync"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. This support multiple initialization issue get
// placing metrics into the context.
var (
	m  *Metrics
	mu sync.Mutex
)

// =============================================================================

// Metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently. No extra abstraction is required.
type Metrics struct {
	Goroutines *expvar.Int
	Requests   *expvar.Int
	Errors     *expvar.Int
	Panics     *expvar.Int
}

// init constructs the metrics value that will be used to capture metrics.
// The metrics value is stored in a package level variable incase this
// function is called more than once. We don't want multiple instances.
func init() {
	mu.Lock()
	defer mu.Unlock()

	// If the metrics value was not constructed yet, then
	// perform the construction.
	if m == nil {
		m = &Metrics{
			Goroutines: expvar.NewInt("goroutines"),
			Requests:   expvar.NewInt("requests"),
			Errors:     expvar.NewInt("errors"),
			Panics:     expvar.NewInt("panics"),
		}
	}
}

// =============================================================================

// This code assumes that the metrics value constructed in main will be
// present in each request. For this codebase, that is done by the
// metrics middleware.

// ctxKeyMetric represents the type of value for the context key.
type ctxKey int

// key is how metric values are stored/retrieved.
const key ctxKey = 1

// =============================================================================

// Set sets the metrics data into the context.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

// Add more of these functions when a metric needs to be collected in
// different parts of the codebase. This will keep this package the
// central authority for metrics and metrics won't get lost.

// AddGoroutines increments the goroutines metric by 1.
func AddGoroutines(ctx context.Context) {
	if v, ok := ctx.Value(key).(*Metrics); ok {
		if v.Goroutines.Value()%100 == 0 {
			v.Goroutines.Add(1)
		}
	}
}

// AddRequests increments the request metric by 1.
func AddRequests(ctx context.Context) {
	if v, ok := ctx.Value(key).(*Metrics); ok {
		v.Requests.Add(1)
	}
}

// AddErrors increments the errors metric by 1.
func AddErrors(ctx context.Context) {
	if v, ok := ctx.Value(key).(*Metrics); ok {
		v.Errors.Add(1)
	}
}

// AddPanics increments the panics metric by 1.
func AddPanics(ctx context.Context) {
	if v, ok := ctx.Value(key).(*Metrics); ok {
		v.Panics.Add(1)
	}
}
