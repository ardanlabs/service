// Package userdb contains user related CRUD functionality.
package userdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/ardanlabs/service/business/sys/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log    *zap.SugaredLogger
	db     sqlx.ExtContext
	inTran bool
}

// NewStore constructs the api for data access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s *Store) WithinTran(ctx context.Context, fn func(s user.Storer) error) error {
	if s.inTran {
		return fn(s)
	}

	f := func(tx *sqlx.Tx) error {
		s := &Store{
			log:    s.log,
			db:     tx,
			inTran: true,
		}
		return fn(s)
	}

	return database.WithinTran(ctx, s.log, s.db.(*sqlx.DB), f)
}

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr user.User) error {
	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, enabled, date_created, date_updated)
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :enabled, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, database.ErrDBDuplicatedEntry) {
			return fmt.Errorf("create: %w", user.ErrUniqueEmail)
		}
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(ctx context.Context, usr user.User) error {
	const q = `
	UPDATE
		users
	SET 
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, database.ErrDBDuplicatedEntry) {
			return user.ErrUniqueEmail
		}
		return fmt.Errorf("updating userID[%s]: %w", usr.ID, err)
	}

	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, userID string) error {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting userID[%s]: %w", userID, err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter user.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]user.User, error) {
	data := struct {
		ID          string `db:"id"`
		Name        string `db:"name"`
		Email       string `db:"email"`
		Offset      int    `db:"offset"`
		RowsPerPage int    `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString("SELECT * FROM users ")

	var wc []string

	if filter.ID != nil {
		data.ID = *filter.ID
		wc = append(wc, "id = :id")
	}

	if filter.Name != nil {
		data.Name = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Email != nil {
		data.Email = *filter.Email
		wc = append(wc, "email = :email")
	}

	if len(wc) > 0 {
		buf.WriteString("WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}

	buf.WriteString(" ORDER BY ")
	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var usrs []dbUser
	if err := database.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &usrs); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return toCoreUserSlice(usrs), nil
}

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID string) (user.User, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE 
		user_id = :user_id`

	var usr dbUser
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("selecting userID[%q]: %w", userID, err)
	}

	return toCoreUser(usr), nil
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email string) (user.User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = :email`

	var usr dbUser
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("selecting email[%q]: %w", email, err)
	}

	return toCoreUser(usr), nil
}
