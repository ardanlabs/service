// Package product contains product related CRUD functionality.
package product

import (
	"context"
	"database/sql"
	"time"

	"github.com/ardanlabs/service/business/auth"
	"github.com/ardanlabs/service/business/data"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
)

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func Create(ctx context.Context, db *sqlx.DB, user auth.Claims, np data.NewProduct, now time.Time) (data.Product, error) {
	ctx, span := global.Tracer("service").Start(ctx, "internal.data.product.create")
	defer span.End()

	p := data.Product{
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

	if _, err := db.ExecContext(ctx, q, p.ID, p.UserID, p.Name, p.Cost, p.Quantity, p.DateCreated, p.DateUpdated); err != nil {
		return data.Product{}, errors.Wrap(err, "inserting product")
	}

	return p, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func Update(ctx context.Context, db *sqlx.DB, user auth.Claims, id string, update data.UpdateProduct, now time.Time) error {
	ctx, span := global.Tracer("service").Start(ctx, "internal.data.product.update")
	defer span.End()

	p, err := One(ctx, db, id)
	if err != nil {
		return err
	}

	// If you are not an admin and looking to retrieve someone elses product.
	if !user.HasRole(auth.RoleAdmin) && p.UserID != user.Subject {
		return data.ErrForbidden
	}

	if update.Name != nil {
		p.Name = *update.Name
	}
	if update.Cost != nil {
		p.Cost = *update.Cost
	}
	if update.Quantity != nil {
		p.Quantity = *update.Quantity
	}
	p.DateUpdated = now

	const q = `UPDATE products SET
		"name" = $2,
		"cost" = $3,
		"quantity" = $4,
		"date_updated" = $5
		WHERE product_id = $1`

	if _, err = db.ExecContext(ctx, q, id, p.Name, p.Cost, p.Quantity, p.DateUpdated); err != nil {
		return errors.Wrap(err, "updating product")
	}

	return nil
}

// Delete removes the product identified by a given ID.
func Delete(ctx context.Context, db *sqlx.DB, id string) error {
	ctx, span := global.Tracer("service").Start(ctx, "internal.data.product.delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return data.ErrInvalidID
	}

	const q = `DELETE FROM products WHERE product_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting product %s", id)
	}

	return nil
}

// List gets all Products from the database.
func List(ctx context.Context, db *sqlx.DB) ([]data.Product, error) {
	ctx, span := global.Tracer("service").Start(ctx, "internal.data.product.list")
	defer span.End()

	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity) ,0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		GROUP BY p.product_id`

	products := []data.Product{}
	if err := db.SelectContext(ctx, &products, q); err != nil {
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// One finds the product identified by a given ID.
func One(ctx context.Context, db *sqlx.DB, id string) (data.Product, error) {
	ctx, span := global.Tracer("service").Start(ctx, "internal.data.product.one")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return data.Product{}, data.ErrInvalidID
	}

	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity), 0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		WHERE p.product_id = $1
		GROUP BY p.product_id`

	var p data.Product
	if err := db.GetContext(ctx, &p, q, id); err != nil {
		if err == sql.ErrNoRows {
			return data.Product{}, data.ErrNotFound
		}
		return data.Product{}, errors.Wrap(err, "selecting single product")
	}

	return p, nil
}
