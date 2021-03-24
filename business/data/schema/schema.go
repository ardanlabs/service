// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	"bufio"
	"context"
	_ "embed" // Calls init function.
	"strconv"
	"strings"

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

	d := darwin.New(driver, parseMigrations(schemaDoc))
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

func parseMigrations(s string) []darwin.Migration {
	var migs []darwin.Migration

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)

	var mig darwin.Migration
	var script string
	for scanner.Scan() {
		v := strings.ToLower(scanner.Text())
		switch {
		case len(v) >= 5 && (v[:6] == "-- ver" || v[:5] == "--ver"):
			mig.Script = script
			migs = append(migs, mig)

			mig = darwin.Migration{}
			script = ""

			f, err := strconv.ParseFloat(strings.TrimSpace(v[11:]), 64)
			if err != nil {
				return nil
			}
			mig.Version = f

		case len(v) >= 5 && (v[:6] == "-- des" || v[:5] == "--des"):
			mig.Description = strings.TrimSpace(v[15:])

		default:
			script += v + "\n"
		}
	}

	mig.Script = script
	migs = append(migs, mig)

	return migs[1:]
}
