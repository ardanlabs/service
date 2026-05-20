package dialect

import "bytes"

// SQLite is the Dialect for SQLite. It uses the LIMIT / OFFSET pagination
// form, which is also what MySQL and modern Postgres accept. It exists as
// a second, deliberately different dialect so the engine-specific seam
// in the store layer is visible and exercised.
type SQLite struct{}

// Name implements Dialect.
func (SQLite) Name() string {
	return "sqlite"
}

// Paginate implements Dialect.
func (SQLite) Paginate(buf *bytes.Buffer) {
	buf.WriteString(" LIMIT :rows_per_page OFFSET :offset")
}
