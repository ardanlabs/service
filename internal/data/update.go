package data

import (
	"context"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/trace"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type update struct{}

// Update contains the data api for anything related to updating data.
var Update update

// User replaces a user document in the database.
func (update) User(ctx context.Context, claims auth.Claims, db *sqlx.DB, id string, upd UpdateUser, now time.Time) error {
	ctx = trace.NewSpan(ctx, "internal.data.update.user")

	u, err := Retrieve.User.One(ctx, claims, db, id)
	if err != nil {
		return err
	}

	if upd.Name != nil {
		u.Name = *upd.Name
	}
	if upd.Email != nil {
		u.Email = *upd.Email
	}
	if upd.Roles != nil {
		u.Roles = upd.Roles
	}
	if upd.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*upd.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		u.PasswordHash = pw
	}

	u.DateUpdated = now

	const q = `UPDATE users SET
		"name" = $2,
		"email" = $3,
		"roles" = $4,
		"password_hash" = $5,
		"date_updated" = $6
		WHERE user_id = $1`
	_, err = db.ExecContext(ctx, q, id, u.Name, u.Email, u.Roles, u.PasswordHash, u.DateUpdated)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Product modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (update) Product(ctx context.Context, db *sqlx.DB, user auth.Claims, id string, update UpdateProduct, now time.Time) error {
	ctx = trace.NewSpan(ctx, "internal.data.update.product")

	p, err := Retrieve.Product.One(ctx, db, id)
	if err != nil {
		return err
	}

	// If you do not have the admin role ...
	// and you are not the owner of this product ...
	// then get outta here!
	if !user.HasRole(auth.RoleAdmin) && p.UserID != user.Subject {
		return ErrForbidden
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
	_, err = db.ExecContext(ctx, q, id,
		p.Name, p.Cost,
		p.Quantity, p.DateUpdated,
	)
	if err != nil {
		return errors.Wrap(err, "updating product")
	}

	return nil
}
