// Package productdb contains product related CRUD functionality.
package productdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/business/sys/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for product database access.
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

// OrderingFields returns the field set for this store.
func (s *Store) OrderingFields() order.FieldSet {
	return orderingFields
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (s *Store) Create(ctx context.Context, prd product.Product) error {
	const q = `
	INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
	VALUES
		(:product_id, :user_id, :name, :cost, :quantity, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(prd)); err != nil {
		return fmt.Errorf("inserting product: %w", err)
	}

	return nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (s *Store) Update(ctx context.Context, prd product.Product) error {
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

	if err := database.NamedExecContext(ctx, s.log, s.db, q, toDBProduct(prd)); err != nil {
		return fmt.Errorf("updating product productID[%s]: %w", prd.ID, err)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (s *Store) Delete(ctx context.Context, prd product.Product) error {
	data := struct {
		ProductID string `db:"product_id"`
	}{
		ProductID: prd.ID.String(),
	}

	const q = `
	DELETE FROM
		products
	WHERE
		product_id = :product_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting product productID[%s]: %w", prd.ID, err)
	}

	return nil
}

// Query gets all Products from the database.
func (s *Store) Query(ctx context.Context, filter product.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]product.Product, error) {
	data := struct {
		ID          string `db:"product_id"`
		Name        string `db:"name"`
		Cost        int    `db:"cost"`
		Quantity    int    `db:"quantity"`
		Offset      int    `db:"offset"`
		RowsPerPage int    `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	orderByClause, err := orderBy.Clause()
	if err != nil {
		return nil, err
	}

	var wc []string
	if filter.ID != nil {
		data.ID = (*filter.ID).String()
		wc = append(wc, "product_id = :product_id")
	}
	if filter.Name != nil {
		data.Name = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name LIKE :name")
	}
	if filter.Cost != nil {
		data.Cost = *filter.Cost
		wc = append(wc, "cost = :cost")
	}
	if filter.Quantity != nil {
		data.Quantity = *filter.Quantity
		wc = append(wc, "quantity = :quantity")
	}

	const q = `
	SELECT
		p.*,
		COALESCE(SUM(s.quantity) ,0) AS sold,
		COALESCE(SUM(s.paid), 0) AS revenue
	FROM
		products AS p
	LEFT JOIN
		sales AS s ON p.product_id = s.product_id
	`
	buf := bytes.NewBufferString(q)

	if len(wc) > 0 {
		buf.WriteString("WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
	buf.WriteString(" GROUP BY p.product_id ")
	buf.WriteString(" ORDER BY ")
	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var prds []dbProduct
	if err := database.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &prds); err != nil {
		return nil, fmt.Errorf("selecting products: %w", err)
	}

	return toCoreProductSlice(prds), nil
}

// QueryByID finds the product identified by a given ID.
func (s *Store) QueryByID(ctx context.Context, productID uuid.UUID) (product.Product, error) {
	data := struct {
		ProductID string `db:"product_id"`
	}{
		ProductID: productID.String(),
	}

	const q = `
	SELECT
		p.*,
		COALESCE(SUM(s.quantity), 0) AS sold,
		COALESCE(SUM(s.paid), 0) AS revenue
	FROM
		products AS p
	LEFT JOIN
		sales AS s ON p.product_id = s.product_id
	WHERE
		p.product_id = :product_id
	GROUP BY
		p.product_id`

	var prd dbProduct
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &prd); err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return product.Product{}, product.ErrNotFound
		}
		return product.Product{}, fmt.Errorf("selecting product productID[%q]: %w", productID, err)
	}

	return toCoreProduct(prd), nil
}

// QueryByUserID finds the product identified by a given User ID.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]product.Product, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID.String(),
	}

	const q = `
	SELECT
		p.*,
		COALESCE(SUM(s.quantity), 0) AS sold,
		COALESCE(SUM(s.paid), 0) AS revenue
	FROM
		products AS p
	LEFT JOIN
		sales AS s ON p.product_id = s.product_id
	WHERE
		p.user_id = :user_id
	GROUP BY
		p.product_id`

	var prds []dbProduct
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, data, &prds); err != nil {
		return nil, fmt.Errorf("selecting products userID[%s]: %w", userID, err)
	}

	return toCoreProductSlice(prds), nil
}
