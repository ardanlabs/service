// Package productsqlite contains the SQLite implementation of the product
// store. It exists primarily as a second engine implementation so the
// pattern of "share the engine-agnostic helpers, own the SQL per engine"
// is visible side-by-side with productpg.
//
// The two stores look almost identical; the meaningful differences are:
//
//   - The dialect plugged into the Store is dialect.SQLite, so the
//     pagination clause becomes LIMIT/OFFSET instead of OFFSET/FETCH.
//   - The Update statement omits the double-quoted identifiers used by
//     productpg. This is a deliberate, function-level deviation that
//     demonstrates the "you can still customize per function" property:
//     each engine owns its own SQL strings.
package productsqlite

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/productbus/stores/commondb"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/business/sdk/sqldb/dialect"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for product database access on SQLite.
type Store struct {
	log     *logger.Logger
	db      sqlx.ExtContext
	dialect dialect.Dialect
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log:     log,
		db:      db,
		dialect: dialect.SQLite{},
	}
}

// NewWithTx constructs a new Store value replacing the sqlx DB value with
// a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (productbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log:     s.log,
		db:      ec,
		dialect: s.dialect,
	}

	return &store, nil
}

// Create adds a Product to the database.
func (s *Store) Create(ctx context.Context, prd productbus.Product) error {
	const q = `
	INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
	VALUES
		(:product_id, :user_id, :name, :cost, :quantity, :date_created, :date_updated)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, commondb.ToDBProduct(prd)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about a product. SQLite does not require identifier
// quoting for any of these column names, so this statement reads slightly
// cleaner than its Postgres counterpart. This is the demonstrative example
// of per-function customization: each engine writes the SQL that suits it.
func (s *Store) Update(ctx context.Context, prd productbus.Product) error {
	const q = `
	UPDATE
		products
	SET
		name = :name,
		cost = :cost,
		quantity = :quantity,
		date_updated = :date_updated
	WHERE
		product_id = :product_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, commondb.ToDBProduct(prd)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (s *Store) Delete(ctx context.Context, prd productbus.Product) error {
	data := struct {
		ID string `db:"product_id"`
	}{
		ID: prd.ID.String(),
	}

	const q = `
	DELETE FROM
		products
	WHERE
		product_id = :product_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query gets all Products from the database matching the filter.
func (s *Store) Query(ctx context.Context, filter productbus.QueryFilter, orderBy order.By, pg page.Page) ([]productbus.Product, error) {
	data := map[string]any{
		"offset":        (pg.Number() - 1) * pg.RowsPerPage(),
		"rows_per_page": pg.RowsPerPage(),
	}

	const q = `
	SELECT
	    product_id, user_id, name, cost, quantity, date_created, date_updated
	FROM
		products`

	buf := bytes.NewBufferString(q)
	commondb.ApplyFilter(filter, data, buf)

	clause, err := commondb.OrderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(clause)
	s.dialect.Paginate(buf)

	var dbPrds []commondb.ProductDB
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPrds); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return commondb.ToBusProducts(dbPrds)
}

// Count returns the total number of products in the DB matching the filter.
func (s *Store) Count(ctx context.Context, filter productbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		count(1)
	FROM
		products`

	buf := bytes.NewBufferString(q)
	commondb.ApplyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// QueryByID finds the product identified by a given ID.
func (s *Store) QueryByID(ctx context.Context, productID uuid.UUID) (productbus.Product, error) {
	data := struct {
		ID string `db:"product_id"`
	}{
		ID: productID.String(),
	}

	const q = `
	SELECT
	    product_id, user_id, name, cost, quantity, date_created, date_updated
	FROM
		products
	WHERE
		product_id = :product_id`

	var dbPrd commondb.ProductDB
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPrd); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return productbus.Product{}, fmt.Errorf("db: %w", productbus.ErrNotFound)
		}
		return productbus.Product{}, fmt.Errorf("db: %w", err)
	}

	return commondb.ToBusProduct(dbPrd)
}

// QueryByUserID finds products by their owning user ID.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]productbus.Product, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
	    product_id, user_id, name, cost, quantity, date_created, date_updated
	FROM
		products
	WHERE
		user_id = :user_id`

	var dbPrds []commondb.ProductDB
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbPrds); err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return commondb.ToBusProducts(dbPrds)
}
