// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	"context"
	_ "embed" // Calls init function.

	"github.com/ardanlabs/darwin"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var (
	//go:embed sql/schema.sql
	schemaDoc string

	//go:embed sql/seed.sql
	seedDoc string

	//go:embed sql/delete.sql
	deleteDoc string
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return errors.Wrap(err, "status check database")
	}

	driver, err := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	if err != nil {
		return errors.Wrap(err, "construct darwin driver")
	}

	d := darwin.New(driver, darwin.ParseMigrations(schemaDoc))
	return d.Migrate()
}

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return errors.Wrap(err, "status check database")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seedDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// DeleteAll runs the set of Drop-table queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func DeleteAll(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
