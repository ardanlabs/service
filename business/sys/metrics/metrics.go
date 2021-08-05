// Package metrics cosntructs the metrics the application will track.
package metrics

import (
	"context"
	"expvar"
	"sync"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. This is never accessed directly, it's just
// here to maintain the single instance in case the New function
// is called more than once. This is possible with testing.
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

// New constructs the metrics value that will be used to capture metrics.
// The metrics value is stored in a package level variable incase this
// function is called more than once. We don't want multiple instances.
func New() *Metrics {
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
	return m
}

// =============================================================================

// This code assumes that the metrics value constructed in main will be
// present in each request. For this codebase, that is done by the
// metrics middleware.

// ctxKeyMetric represents the type of value for the context key.
type ctxKey int

// Key is how metric values are stored/retrieved.
const Key ctxKey = 1

// =============================================================================

// Add more of these functions when a metric needs to be collected in
// different parts of the codebase. This will keep this package the
// central authority for metrics and metrics won't get lost.
//
// You could also pass the metrics value around the program as well.
// Since this is for debugging, managing, and maintain the app, having
// it in the context is fine.

// AddPanics increments the panics metric by 1.
func AddPanics(ctx context.Context) {
	if v, ok := ctx.Value(Key).(*Metrics); ok {
		v.Panics.Add(1)
	}
}
