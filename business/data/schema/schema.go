// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	// Used to embed schema.sql into the sql variable.
	_ "embed"

	"bufio"
	"strconv"
	"strings"

	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

//go:embed schema.sql
var sql string

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sqlx.DB) error {
	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	d := darwin.New(driver, parseMigrations(sql), nil)
	return d.Migrate()
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

			f, err := strconv.ParseFloat(strings.Trim(v[11:], " "), 64)
			if err != nil {
				return nil
			}
			mig.Version = f

		case len(v) >= 5 && (v[:6] == "-- des" || v[:5] == "--des"):
			mig.Description = strings.Trim(v[15:], " ")

		default:
			script += v + "\n"
		}
	}

	mig.Script = script
	migs = append(migs, mig)

	return migs[1:]
}
