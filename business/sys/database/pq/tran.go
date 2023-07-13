package db

import (
	"fmt"

	"github.com/ardanlabs/service/business/sys/database"
	"github.com/jmoiron/sqlx"
)

// dbBeginner implements the core coreTransaction interface,
type dbBeginner struct {
	sqlxDB *sqlx.DB
}

// NewBeginner constructs a value that implements the database
// beginner interface.
func NewBeginner(sqlxDB *sqlx.DB) database.Beginner {
	return &dbBeginner{
		sqlxDB: sqlxDB,
	}
}

// Begin start a transaction and returns a value that implements
// the core transactor interface.
func (db *dbBeginner) Begin() (database.Transaction, error) {
	return db.sqlxDB.Beginx()
}

// GetExtContext is a helper function that extracts the sqlx value
// from the core transactor interface for transactional use.
func GetExtContext(tx database.Transaction) (sqlx.ExtContext, error) {
	ec, ok := tx.(sqlx.ExtContext)
	if !ok {
		return nil, fmt.Errorf("Transactor(%T) not of a type *sql.Tx", tx)
	}

	return ec, nil
}
