package database

import (
	"fmt"

	"github.com/ardanlabs/service/business/sys/core"
	"github.com/jmoiron/sqlx"
)

type CoreDB struct {
	sqlxDB *sqlx.DB
}

func NewCoreDB(sqlxDB *sqlx.DB) *CoreDB {
	return &CoreDB{
		sqlxDB: sqlxDB,
	}
}

func (db *CoreDB) Begin() (core.Transactor, error) {
	return db.sqlxDB.Beginx()
}

func GetExtContext(tr core.Transactor) (sqlx.ExtContext, error) {
	tx, ok := tr.(sqlx.ExtContext)
	if !ok {
		return nil, fmt.Errorf("Transactor(%T) not of a type *sql.Tx", tr)
	}

	return tx, nil
}
