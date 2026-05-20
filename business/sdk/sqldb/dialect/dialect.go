// Package dialect provides cross-cutting helpers that vary between SQL
// engines but are otherwise reusable by every domain store. A store wires
// in one implementation (Postgres, SQLite, ...) and delegates the small set
// of engine-specific decisions to it.
//
// The set of decisions encapsulated here is intentionally narrow. Only add
// to the interface when a real, second engine forces a difference.
package dialect

import "bytes"

// Dialect describes the engine-specific behavior a store needs in order
// to compose portable SQL from shared fragments.
type Dialect interface {

	// Name reports a short identifier for the engine (for logging only).
	Name() string

	// Paginate appends a pagination clause to buf. The clause must consume
	// the named bind variables ":offset" and ":rows_per_page" already
	// supplied by the caller in the parameter map.
	Paginate(buf *bytes.Buffer)
}
