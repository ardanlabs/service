// Package metrics cosntructs the metrics the application will track.
package metrics

import "expvar"

// ctxKeyMetric represents the type of value for the context key.
type ctxKey int

// Key is how metric values are stored/retrieved.
const Key ctxKey = 1

// Metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently. No extra abstraction is required.
type Metrics struct {
	Goroutines *expvar.Int
	Requets    *expvar.Int
	Errors     *expvar.Int
	Panics     *expvar.Int
}

// New constructs the metrics that will be tracked.
func New() *Metrics {
	return &Metrics{
		Goroutines: expvar.NewInt("goroutines"),
		Requets:    expvar.NewInt("requests"),
		Errors:     expvar.NewInt("errors"),
		Panics:     expvar.NewInt("panics"),
	}
}
