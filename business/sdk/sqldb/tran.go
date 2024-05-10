package sqldb

import (
	"fmt"

	"github.com/ardanlabs/service/business/sdk/transaction"
	"github.com/jmoiron/sqlx"
)

// DBBeginner implements the Beginner interface,
type DBBeginner struct {
	sqlxDB *sqlx.DB
}

// NewBeginner constructs a value that implements the beginner interface.
func NewBeginner(sqlxDB *sqlx.DB) *DBBeginner {
	return &DBBeginner{
		sqlxDB: sqlxDB,
	}
}

// Begin implements the Beginner interface and returns a concrete value that
// implements the CommitRollbacker interface.
func (db *DBBeginner) Begin() (transaction.CommitRollbacker, error) {
	return db.sqlxDB.Beginx()
}

// GetExtContext is a helper function that extracts the sqlx value
// from the domain transactor interface for transactional use.
func GetExtContext(tx transaction.CommitRollbacker) (sqlx.ExtContext, error) {
	ec, ok := tx.(sqlx.ExtContext)
	if !ok {
		return nil, fmt.Errorf("Transactor(%T) not of a type *sql.Tx", tx)
	}

	return ec, nil
}
