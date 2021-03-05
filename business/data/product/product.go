// Package product contains product related CRUD functionality.
package product

import (
	"context"
	"log"
	"time"

	"github.com/ardanlabs/service/business/auth"
	"github.com/ardanlabs/service/business/validate"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrNotFound is used when a specific Product is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// Product manages the set of API's for product access.
type Product struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Product for api access.
func New(log *log.Logger, db *sqlx.DB) Product {
	return Product{
		log: log,
		db:  db,
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (p Product) Create(ctx context.Context, traceID string, claims auth.Claims, np NewProduct, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.create")
	defer span.End()

	if err := validate.Check(np); err != nil {
		return Info{}, errors.Wrap(err, "validating data")
	}

	prd := Info{
		ID:          validate.GenerateID(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      claims.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `
	INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
	VALUES
		(:product_id, :user_id, :name, :cost, :quantity, :date_created, :date_updated)`

	p.log.Printf("%s: %s: %s", traceID, "product.Create",
		database.Log(q, prd),
	)

	if _, err := p.db.NamedExecContext(ctx, q, prd); err != nil {
		return Info{}, errors.Wrap(err, "inserting product")
	}

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (p Product) Update(ctx context.Context, traceID string, claims auth.Claims, productID string, up UpdateProduct, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.update")
	defer span.End()

	if err := validate.CheckID(productID); err != nil {
		return ErrInvalidID
	}
	if err := validate.Check(up); err != nil {
		return errors.Wrap(err, "validating data")
	}

	prd, err := p.QueryByID(ctx, traceID, productID)
	if err != nil {
		return errors.Wrap(err, "updating product")
	}

	// If you are not an admin and looking to retrieve someone elses product.
	if !claims.Authorized(auth.RoleAdmin) && prd.UserID != claims.Subject {
		return ErrForbidden
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

	p.log.Printf("%s: %s: %s", traceID, "product.Update",
		database.Log(q, prd),
	)

	if _, err := p.db.NamedExecContext(ctx, q, prd); err != nil {
		return errors.Wrapf(err, "updating product %s", prd.ID)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (p Product) Delete(ctx context.Context, traceID string, claims auth.Claims, productID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.delete")
	defer span.End()

	if err := validate.CheckID(productID); err != nil {
		return ErrInvalidID
	}

	// If you are not an admin.
	if !claims.Authorized(auth.RoleAdmin) {
		return ErrForbidden
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

	p.log.Printf("%s: %s: %s", traceID, "product.Delete",
		database.Log(q, data),
	)

	if _, err := p.db.NamedExecContext(ctx, q, data); err != nil {
		return errors.Wrapf(err, "deleting product %s", data.ProductID)
	}

	return nil
}

// Query gets all Products from the database.
func (p Product) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.query")
	defer span.End()

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

	p.log.Printf("%s: %s: %s", traceID, "product.Query",
		database.Log(q, data),
	)

	var products []Info
	if err := database.NamedQuerySlice(ctx, p.db, q, data, &products); err != nil {
		if err == database.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// QueryByID finds the product identified by a given ID.
func (p Product) QueryByID(ctx context.Context, traceID string, productID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.querybyid")
	defer span.End()

	if err := validate.CheckID(productID); err != nil {
		return Info{}, ErrInvalidID
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

	p.log.Printf("%s: %s: %s", traceID, "product.QueryByID",
		database.Log(q, data),
	)

	var prd Info
	if err := database.NamedQueryStruct(ctx, p.db, q, data, &prd); err != nil {
		if err == database.ErrNotFound {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", data.ProductID)
	}

	return prd, nil
}
