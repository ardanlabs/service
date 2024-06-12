// Package productdb contains product related CRUD functionality.
package productdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for product database access.
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

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (productbus.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create adds a Product to the sqldb. It returns the created Product with
// fields like ID and DateCreated populated.
func (s *Store) Create(ctx context.Context, prd productbus.Product) error {
	const q = `
	INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
	VALUES
		(:product_id, :user_id, :name, :cost, :quantity, :date_created, :date_updated)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(prd)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update modifies data about a productbus. It will error if the specified ID is
// invalid or does not reference an existing productbus.
func (s *Store) Update(ctx context.Context, prd productbus.Product) error {
	const q = `
	UPDATE
		products
	SET
		"name" = :name,
		"cost" = :cost,
		"quantity" = :quantity,
		"date_updated" = :date_updated
	WHERE
		product_id = :product_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(prd)); err != nil {
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

// Query gets all Products from the database.
func (s *Store) Query(ctx context.Context, filter productbus.QueryFilter, orderBy order.By, page page.Page) ([]productbus.Product, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
	}

	const q = `
	SELECT
	    product_id, user_id, name, cost, quantity, date_created, date_updated
	FROM
		products`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbPrds []product
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbPrds); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusProducts(dbPrds)
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter productbus.QueryFilter) (int, error) {
	data := map[string]any{}

	const q = `
	SELECT
		count(1)
	FROM
		products`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count   int `db:"count"`
		Sold    int `db:"sold"`
		Revenue int `db:"revenue"`
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

	var dbPrd product
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPrd); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return productbus.Product{}, fmt.Errorf("db: %w", productbus.ErrNotFound)
		}
		return productbus.Product{}, fmt.Errorf("db: %w", err)
	}

	return toBusProduct(dbPrd)
}

// QueryByUserID finds the product identified by a given User ID.
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

	var dbPrds []product
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbPrds); err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return toBusProducts(dbPrds)
}
