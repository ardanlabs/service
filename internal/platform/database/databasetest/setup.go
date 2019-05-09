package databasetest

import (
	"testing"
	"time"

	"github.com/ardanlabs/garagesale/internal/platform/database"
	"github.com/ardanlabs/garagesale/internal/platform/database/schema"
	"github.com/jmoiron/sqlx"
)

// Setup creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty.
//
// It does not return errors as this intended for testing only. Instead it will
// call Fatal on the provided testing.T if anything goes wrong.
//
// It returns the database to use as well as a function to call at the end of
// the test.
func Setup(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	c := startContainer(t)

	cfg := database.Config{
		DisableTLS: true,
		Host:       c.Host,
		Name:       "postgres",
		Password:   "postgres",
		User:       "postgres",
	}

	db, err := database.Open(cfg)
	if err != nil {
		t.Fatalf("opening database connection: %v", err)
	}

	t.Log("waiting for database to be ready")

	// Wait for the database to be ready. Wait 100ms longer between each attempt.
	// Do not try more than 20 times.
	var pingError error
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		stopContainer(t, c)
		t.Fatalf("waiting for database to be ready: %v", pingError)
	}

	if err := schema.Migrate(db); err != nil {
		stopContainer(t, c)
		t.Fatalf("migrating: %s", err)
	}

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		db.Close()
		stopContainer(t, c)
	}

	return db, teardown
}
