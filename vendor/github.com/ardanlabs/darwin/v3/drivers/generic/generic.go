// Package generic implements a generic driver.
package generic

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ardanlabs/darwin/v3"
)

// Dialect is used to support multiple databases by returning proper SQL.
type Dialect interface {
	CreateTableSQL() string
	InsertSQL() string
	UpdateChecksumSQL() string
	AllSQL() string
}

// =============================================================================

// Driver is the default Driver, it can be configured to any database.
type Driver struct {
	DB      *sql.DB
	Dialect Dialect
}

// New creates a new GenericDriver configured with db and dialect.
func New(db *sql.DB, dialect Dialect) (*Driver, error) {
	if db == nil {
		return nil, errors.New("darwin: db is nil")
	}

	if dialect == nil {
		return nil, errors.New("darwin: dialectis nil")
	}

	return &Driver{DB: db, Dialect: dialect}, nil
}

// Create create the table darwin_migrations if necessary.
func (d *Driver) Create() error {
	f := func(tx *sql.Tx) error {
		_, err := tx.Exec(d.Dialect.CreateTableSQL())
		return err
	}
	return transaction(d.DB, f)
}

// Insert insert a migration entry into database.
func (d *Driver) Insert(e darwin.MigrationRecord) error {
	f := func(tx *sql.Tx) error {
		_, err := tx.Exec(d.Dialect.InsertSQL(),
			e.Version,
			e.Description,
			e.Checksum,
			e.AppliedAt.Unix(),
			e.ExecutionTime,
		)
		return err
	}
	return transaction(d.DB, f)
}

// UpdateChecksum updates all the checksums for the migration into the database.
func (d *Driver) UpdateChecksum(checksum string, version float64) error {
	f := func(tx *sql.Tx) error {
		_, err := tx.Exec(d.Dialect.UpdateChecksumSQL(),
			checksum,
			version,
		)
		return err
	}
	return transaction(d.DB, f)
}

// All returns all migrations applied.
func (d *Driver) All() ([]darwin.MigrationRecord, error) {
	rows, err := d.DB.Query(d.Dialect.AllSQL())
	if err != nil {
		return []darwin.MigrationRecord{}, err
	}

	var entries []darwin.MigrationRecord
	for rows.Next() {
		var (
			version       float64
			description   string
			checksum      string
			appliedAt     int64
			executionTime float64
		)

		rows.Scan(
			&version,
			&description,
			&checksum,
			&appliedAt,
			&executionTime,
		)

		entry := darwin.MigrationRecord{
			Version:       version,
			Description:   description,
			Checksum:      checksum,
			AppliedAt:     time.Unix(appliedAt, 0),
			ExecutionTime: time.Duration(executionTime),
		}

		entries = append(entries, entry)
	}

	rows.Close()

	// The PGX driver did not provide float values that we absolute at the
	// precision of a 64 bit float. This fixes that, but does restrict the
	// versioning to 5 decimal points.
	for i := range entries {
		f := fmt.Sprintf("%5f", entries[i].Version)

		var err error
		entries[i].Version, err = strconv.ParseFloat(f, 64)
		if err != nil {
			return nil, err
		}
	}

	return entries, nil
}

// Exec execute sql scripts into database.
func (d *Driver) Exec(script string) (time.Duration, error) {
	start := time.Now()

	f := func(tx *sql.Tx) error {
		_, err := tx.Exec(script)
		return err
	}

	return time.Since(start), transaction(d.DB, f)
}

// =============================================================================

// transaction is a utility function to execute the SQL inside a transaction.
// see: http://stackoverflow.com/a/23502629
func transaction(db *sql.DB, f func(*sql.Tx) error) (err error) {
	if db == nil {
		return errors.New("darwin: sql.DB is nil")
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
			}
		}
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	return f(tx)
}
