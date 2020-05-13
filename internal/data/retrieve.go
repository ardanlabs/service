package data

import (
	"context"
	"database/sql"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

type retrieve struct {
	User    retUser
	Product retProduct
}

// Retrieve contains the data api for anything related to retrieving data.
var Retrieve retrieve

type retUser struct{}

// List retrieves a list of existing users from the database.
func (retUser) List(ctx context.Context, db *sqlx.DB) ([]User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.data.retrieve.user.list")
	defer span.End()

	users := []User{}
	const q = `SELECT * FROM users`

	if err := db.SelectContext(ctx, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// One gets the specified user from the database.
func (retUser) One(ctx context.Context, claims auth.Claims, db *sqlx.DB, id string) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.data.retrieve.user.one")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !claims.HasRole(auth.RoleAdmin) && claims.Subject != id {
		return nil, ErrForbidden
	}

	var u User
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := db.GetContext(ctx, &u, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting user %q", id)
	}

	return &u, nil
}

type retProduct struct{}

// List gets all Products from the database.
func (retProduct) List(ctx context.Context, db *sqlx.DB) ([]Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.data.retrieve.product.list")
	defer span.End()

	products := []Product{}
	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity) ,0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		GROUP BY p.product_id`

	if err := db.SelectContext(ctx, &products, q); err != nil {
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// One finds the product identified by a given ID.
func (retProduct) One(ctx context.Context, db *sqlx.DB, id string) (*Product, error) {
	ctx, span := trace.StartSpan(ctx, "internal.data.retrieve.product.one")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var p Product

	const q = `SELECT
			p.*,
			COALESCE(SUM(s.quantity), 0) AS sold,
			COALESCE(SUM(s.paid), 0) AS revenue
		FROM products AS p
		LEFT JOIN sales AS s ON p.product_id = s.product_id
		WHERE p.product_id = $1
		GROUP BY p.product_id`

	if err := db.GetContext(ctx, &p, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrap(err, "selecting single product")
	}

	return &p, nil
}
