// Package product contains product related CRUD functionality.
package product

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/ardanlabs/service/business/auth"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/google/uuid"
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

	prd := Info{
		ID:          uuid.New().String(),
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
		($1, $2, $3, $4, $5, $6, $7)`

	p.log.Printf("%s: %s: %s", traceID, "product.Create",
		database.Log(q, prd.ID, prd.UserID, prd.Name, prd.Cost, prd.Quantity, prd.DateCreated, prd.DateUpdated),
	)

	if _, err := p.db.ExecContext(ctx, q, prd.ID, prd.UserID, prd.Name, prd.Cost, prd.Quantity, prd.DateCreated, prd.DateUpdated); err != nil {
		return Info{}, errors.Wrap(err, "inserting product")
	}

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (p Product) Update(ctx context.Context, traceID string, claims auth.Claims, productID string, up UpdateProduct, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.update")
	defer span.End()

	prd, err := p.QueryByID(ctx, traceID, productID)
	if err != nil {
		return err
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
		"name" = $2,
		"cost" = $3,
		"quantity" = $4,
		"date_updated" = $5
	WHERE
		product_id = $1`

	p.log.Printf("%s: %s: %s", traceID, "product.Update",
		database.Log(q, productID, prd.Name, prd.Cost, prd.Quantity, prd.DateUpdated),
	)

	if _, err = p.db.ExecContext(ctx, q, productID, prd.Name, prd.Cost, prd.Quantity, prd.DateUpdated); err != nil {
		return errors.Wrap(err, "updating product")
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (p Product) Delete(ctx context.Context, traceID string, claims auth.Claims, productID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.delete")
	defer span.End()

	if _, err := uuid.Parse(productID); err != nil {
		return ErrInvalidID
	}

	// If you are not an admin.
	if !claims.Authorized(auth.RoleAdmin) {
		return ErrForbidden
	}

	const q = `
	DELETE FROM
		products
	WHERE
		product_id = $1`

	p.log.Printf("%s: %s: %s", traceID, "product.Delete",
		database.Log(q, productID),
	)

	if _, err := p.db.ExecContext(ctx, q, productID); err != nil {
		return errors.Wrapf(err, "deleting product %s", productID)
	}

	return nil
}

// Query gets all Products from the database.
func (p Product) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.query")
	defer span.End()

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
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`
	offset := (pageNumber - 1) * rowsPerPage

	p.log.Printf("%s: %s: %s", traceID, "product.Query",
		database.Log(q, offset, rowsPerPage),
	)

	products := []Info{}
	if err := p.db.SelectContext(ctx, &products, q, offset, rowsPerPage); err != nil {
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// QueryByID finds the product identified by a given ID.
func (p Product) QueryByID(ctx context.Context, traceID string, productID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.querybyid")
	defer span.End()

	if _, err := uuid.Parse(productID); err != nil {
		return Info{}, ErrInvalidID
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
		p.product_id = $1
	GROUP BY
		p.product_id`

	p.log.Printf("%s: %s: %s", traceID, "product.QueryByID",
		database.Log(q, productID),
	)

	var prd Info
	if err := p.db.GetContext(ctx, &prd, q, productID); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrap(err, "selecting single product")
	}

	return prd, nil
}
