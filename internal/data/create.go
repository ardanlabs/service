package data

import (
	"context"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/trace"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type create struct{}

// Create contains the data api for anything related to creating data.
var Create create

// User inserts a new user into the database.
func (create) User(ctx context.Context, db *sqlx.DB, n NewUser, now time.Time) (*User, error) {
	ctx = trace.NewSpan(ctx, "internal.data.create.user")

	hash, err := bcrypt.GenerateFromPassword([]byte(n.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "generating password hash")
	}

	u := User{
		ID:           uuid.New().String(),
		Name:         n.Name,
		Email:        n.Email,
		PasswordHash: hash,
		Roles:        n.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.PasswordHash, u.Roles, u.DateCreated, u.DateUpdated)
	if err != nil {
		return nil, errors.Wrap(err, "inserting user")
	}

	return &u, nil
}

// Product adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (create) Product(ctx context.Context, db *sqlx.DB, user auth.Claims, np NewProduct, now time.Time) (*Product, error) {
	ctx = trace.NewSpan(ctx, "internal.data.create.product")

	p := Product{
		ID:          uuid.New().String(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      user.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `
		INSERT INTO products
		(product_id, user_id, name, cost, quantity, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.ExecContext(ctx, q,
		p.ID, p.UserID,
		p.Name, p.Cost, p.Quantity,
		p.DateCreated, p.DateUpdated)
	if err != nil {
		return nil, errors.Wrap(err, "inserting product")
	}

	return &p, nil
}
