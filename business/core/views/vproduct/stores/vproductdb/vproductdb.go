// Package vproductdb provides access to the product view.
package vproductdb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/core/views/vproduct"
	"github.com/ardanlabs/service/business/data/sqldb"
	"github.com/ardanlabs/service/business/web/v1/order"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for product view database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Query retrieves a list of existing products from the database.
func (s *Store) Query(ctx context.Context, filter vproduct.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]vproduct.Product, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `
	SELECT
		product_id,
		user_id,
		name,
		cost,
		quantity,
		date_created,
		date_updated,
		user_name
	FROM
		view_products`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dnPrd []dbProduct
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dnPrd); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toCoreProducts(dnPrd), nil
}

// Count returns the total number of products in the DB.
func (s *Store) Count(ctx context.Context, filter vproduct.QueryFilter) (int, error) {
	data := map[string]interface{}{}

	const q = `
	SELECT
		count(1)
	FROM
		view_products`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}
