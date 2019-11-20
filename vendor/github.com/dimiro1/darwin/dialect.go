package darwin

// Dialect is used to use multiple databases
type Dialect interface {
	// CreateTableSQL returns the SQL to create the schema table
	CreateTableSQL() string

	// InsertSQL returns the SQL to insert a new migration in the schema table
	InsertSQL() string

	// AllSQL returns a SQL to get all entries in the table
	AllSQL() string
}
