package dialect

import "bytes"

// Postgres is the Dialect for PostgreSQL. It uses the SQL:2008 standard
// OFFSET / FETCH NEXT pagination form, which is what the current stores
// in this repository emit today.
type Postgres struct{}

// Name implements Dialect.
func (Postgres) Name() string {
	return "postgres"
}

// Paginate implements Dialect.
func (Postgres) Paginate(buf *bytes.Buffer) {
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")
}
