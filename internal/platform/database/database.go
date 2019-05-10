package database

import (
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // The database driver in use.
)

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	DisableTLS bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {

	// Define SSL mode.
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	// Query parameters.
	var q url.Values
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	// Construct url.
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	return sqlx.Open("postgres", u.String())
}
