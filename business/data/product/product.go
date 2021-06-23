// Package product contains product related CRUD functionality.
package product

import (
	"context"
	"time"

	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Store manages the set of API's for product access.
type Store struct {
	log *zap.SugaredLogger
	db  *sqlx.DB
}

// NewStore constructs a product store for api access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) Store {
	return Store{
		log: log,
		db:  db,
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (s Store) Create(ctx context.Context, traceID string, claims auth.Claims, np NewProduct, now time.Time) (Product, error) {
	if err := validate.Check(np); err != nil {
		return Product{}, errors.Wrap(err, "validating data")
	}

	prd := Product{
		ID:          validate.GenerateID(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      claims.Subject,
		DateCreated: now,
		DateUpdated: now,
	}

	const q = `
	INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
	VALUES
		(:product_id, :user_id, :name, :cost, :quantity, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, traceID, q, prd); err != nil {
		return Product{}, errors.Wrap(err, "inserting product")
	}

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (s Store) Update(ctx context.Context, traceID string, claims auth.Claims, productID string, up UpdateProduct, now time.Time) error {
	if err := validate.CheckID(productID); err != nil {
		return database.ErrInvalidID
	}
	if err := validate.Check(up); err != nil {
		return errors.Wrap(err, "validating data")
	}

	prd, err := s.QueryByID(ctx, traceID, productID)
	if err != nil {
		return errors.Wrap(err, "updating product")
	}

	// If you are not an admin and looking to retrieve someone elses product.
	if !claims.Authorized(auth.RoleAdmin) && prd.UserID != claims.Subject {
		return database.ErrForbidden
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

	if err := database.NamedExecContext(ctx, s.log, s.db, traceID, q, prd); err != nil {
		return errors.Wrapf(err, "updating product %s", prd.ID)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (s Store) Delete(ctx context.Context, traceID string, claims auth.Claims, productID string) error {
	if err := validate.CheckID(productID); err != nil {
		return database.ErrInvalidID
	}

	// If you are not an admin.
	if !claims.Authorized(auth.RoleAdmin) {
		return database.ErrForbidden
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

	if err := database.NamedExecContext(ctx, s.log, s.db, traceID, q, data); err != nil {
		return errors.Wrapf(err, "deleting product %s", data.ProductID)
	}

	return nil
}

// Query gets all Products from the database.
func (s Store) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Product, error) {
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

	var products []Product
	if err := database.NamedQuerySlice(ctx, s.log, s.db, traceID, q, data, &products); err != nil {
		if err == database.ErrNotFound {
			return nil, database.ErrNotFound
		}
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// QueryByID finds the product identified by a given ID.
func (s Store) QueryByID(ctx context.Context, traceID string, productID string) (Product, error) {
	if err := validate.CheckID(productID); err != nil {
		return Product{}, database.ErrInvalidID
	}

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

	var prd Product
	if err := database.NamedQueryStruct(ctx, s.log, s.db, traceID, q, data, &prd); err != nil {
		if err == database.ErrNotFound {
			return Product{}, database.ErrNotFound
		}
		return Product{}, errors.Wrapf(err, "selecting product %q", data.ProductID)
	}

	return prd, nil
}
