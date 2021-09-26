// Package dbproduct contains product related CRUD functionality.
package dbproduct

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/sys/database"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Data manages the set of API's for product access.
type Data struct {
	log *zap.SugaredLogger
	db  *sqlx.DB
}

// NewData constructs a data for api access.
func NewData(log *zap.SugaredLogger, db *sqlx.DB) Data {
	return Data{
		log: log,
		db:  db,
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (d Data) Create(ctx context.Context, np DBNewProduct, now time.Time) (DBProduct, error) {
	dbPrd := DBProduct{
		ID:          validate.GenerateID(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      np.UserID,
		DateCreated: now,
		DateUpdated: now,
	}

	const q = `
	INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
	VALUES
		(:product_id, :user_id, :name, :cost, :quantity, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, d.log, d.db, q, dbPrd); err != nil {
		return DBProduct{}, fmt.Errorf("inserting product: %w", err)
	}

	return dbPrd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (d Data) Update(ctx context.Context, productID string, up DBUpdateProduct, now time.Time) error {
	prd, err := d.QueryByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("updating product productID[%s]: %w", productID, err)
	}

	if up.Name != nil {
		prd.Name = *up.Name
	}
	if up.Cost != nil {
		prd.Cost = *up.Cost
	}
	if up.Quantity != nil {
		prd.Quantity = *up.Quantity
	}
	prd.DateUpdated = now

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

	if err := database.NamedExecContext(ctx, d.log, d.db, q, prd); err != nil {
		return fmt.Errorf("updating product productID[%s]: %w", productID, err)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (d Data) Delete(ctx context.Context, productID string) error {
	if err := validate.CheckID(productID); err != nil {
		return validate.ErrInvalidID
	}

	data := struct {
		ProductID string `db:"product_id"`
	}{
		ProductID: productID,
	}

	const q = `
	DELETE FROM
		products
	WHERE
		product_id = :product_id`

	if err := database.NamedExecContext(ctx, d.log, d.db, q, data); err != nil {
		return fmt.Errorf("deleting product productID[%s]: %w", productID, err)
	}

	return nil
}

// Query gets all Products from the database.
func (d Data) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]DBProduct, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
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
	GROUP BY
		p.product_id
	ORDER BY
		user_id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var dbPrds []DBProduct
	if err := database.NamedQuerySlice(ctx, d.log, d.db, q, data, &dbPrds); err != nil {
		if err == database.ErrDBNotFound {
			return nil, validate.ErrNotFound
		}
		return nil, fmt.Errorf("selecting products: %w", err)
	}

	return dbPrds, nil
}

// QueryByID finds the product identified by a given ID.
func (d Data) QueryByID(ctx context.Context, productID string) (DBProduct, error) {
	data := struct {
		ProductID string `db:"product_id"`
	}{
		ProductID: productID,
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

	var dbPrd DBProduct
	if err := database.NamedQueryStruct(ctx, d.log, d.db, q, data, &dbPrd); err != nil {
		if err == database.ErrDBNotFound {
			return DBProduct{}, validate.ErrNotFound
		}
		return DBProduct{}, fmt.Errorf("selecting product productID[%q]: %w", productID, err)
	}

	return dbPrd, nil
}

// QueryByUserID finds the product identified by a given User ID.
func (d Data) QueryByUserID(ctx context.Context, userID string) ([]DBProduct, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
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

	var dbPrds []DBProduct
	if err := database.NamedQuerySlice(ctx, d.log, d.db, q, data, &dbPrds); err != nil {
		if err == database.ErrDBNotFound {
			return nil, validate.ErrNotFound
		}
		return nil, fmt.Errorf("selecting products userID[%s]: %w", userID, err)
	}

	return dbPrds, nil
}
