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
	"go.opentelemetry.io/otel/api/trace"
)

var (
	// ErrNotFound is used when a specific Product is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func Create(ctx context.Context, traceID string, log *log.Logger, db *sqlx.DB, user auth.Claims, np NewProduct, now time.Time) (Product, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.create")
	defer span.End()

	p := Product{
		ID:          uuid.New().String(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      user.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	log.Printf("%s : %s : query : %s", traceID, "product.Create",
		database.Log(q, p.ID, p.UserID, p.Name, p.Cost, p.Quantity, p.DateCreated, p.DateUpdated),
	)

	if _, err := db.ExecContext(ctx, q, p.ID, p.UserID, p.Name, p.Cost, p.Quantity, p.DateCreated, p.DateUpdated); err != nil {
		return Product{}, errors.Wrap(err, "inserting product")
	}

	return p, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func Update(ctx context.Context, traceID string, log *log.Logger, db *sqlx.DB, user auth.Claims, id string, up UpdateProduct, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.update")
	defer span.End()

	p, err := QueryByID(ctx, traceID, log, db, id)
	if err != nil {
		return err
	}

	// If you are not an admin and looking to retrieve someone elses product.
	if !user.HasRole(auth.RoleAdmin) && p.UserID != user.Subject {
		return ErrForbidden
	}

	if up.Name != nil {
		p.Name = *up.Name
	}
	if up.Cost != nil {
		p.Cost = *up.Cost
	}
	if up.Quantity != nil {
		p.Quantity = *up.Quantity
	}
	p.DateUpdated = now

	const q = `UPDATE products SET
		"name" = $2,
		"cost" = $3,
		"quantity" = $4,
		"date_updated" = $5
		WHERE product_id = $1`

	log.Printf("%s : %s : query : %s", traceID, "product.Update",
		database.Log(q, id, p.Name, p.Cost, p.Quantity, p.DateUpdated),
	)

	if _, err = db.ExecContext(ctx, q, id, p.Name, p.Cost, p.Quantity, p.DateUpdated); err != nil {
		return errors.Wrap(err, "updating product")
	}

	return nil
}

// Delete removes the product identified by a given ID.
func Delete(ctx context.Context, traceID string, log *log.Logger, db *sqlx.DB, id string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM products WHERE product_id = $1`

	log.Printf("%s : %s : query : %s", traceID, "product.Delete",
		database.Log(q, id),
	)

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting product %s", id)
	}

	return nil
}

// Query gets all Products from the database.
func Query(ctx context.Context, traceID string, log *log.Logger, db *sqlx.DB) ([]Product, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.query")
	defer span.End()

	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity) ,0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		GROUP BY p.product_id`

	log.Printf("%s : %s : query : %s", traceID, "product.Query",
		database.Log(q),
	)

	products := []Product{}
	if err := db.SelectContext(ctx, &products, q); err != nil {
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// QueryByID finds the product identified by a given ID.
func QueryByID(ctx context.Context, traceID string, log *log.Logger, db *sqlx.DB, id string) (Product, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.product.querybyid")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return Product{}, ErrInvalidID
	}

	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity), 0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		WHERE p.product_id = $1
		GROUP BY p.product_id`

	log.Printf("%s : %s : query : %s", traceID, "product.QueryByID",
		database.Log(q, id),
	)

	var p Product
	if err := db.GetContext(ctx, &p, q, id); err != nil {
		if err == sql.ErrNoRows {
			return Product{}, ErrNotFound
		}
		return Product{}, errors.Wrap(err, "selecting single product")
	}

	return p, nil
}
