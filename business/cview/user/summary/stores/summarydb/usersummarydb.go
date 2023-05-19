// Package summarydb contains product related CRUD functionality.
package summarydb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/cview/user/summary"
	"github.com/ardanlabs/service/business/data/order"
	database "github.com/ardanlabs/service/business/sys/database/pq"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log *zap.SugaredLogger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter summary.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]summary.Summary, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		user_summary`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbSmm []dbSummary
	if err := database.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbSmm); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreSummarySlice(dbSmm), nil
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter summary.QueryFilter) (int, error) {
	data := map[string]interface{}{}

	const q = `
	SELECT
		count(1)
	FROM
		user_summary`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := database.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}
