// Package migrate contains the database schema, migrations and seeding data.
package migrate

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/ardanlabs/darwin/v3"
	"github.com/ardanlabs/darwin/v3/dialects/postgres"
	"github.com/ardanlabs/darwin/v3/drivers/generic"
	"github.com/ardanlabs/service/business/data/sqldb"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/migrate.sql
	migrateDoc string
)

// Migrate attempts to bring the database up to date with the migrations
// defined in this package.
func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := sqldb.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	driver, err := generic.New(db.DB, postgres.Dialect{})
	if err != nil {
		return fmt.Errorf("construct darwin driver: %w", err)
	}

	d := darwin.New(driver, darwin.ParseMigrations(migrateDoc))
	return d.Migrate()
}
