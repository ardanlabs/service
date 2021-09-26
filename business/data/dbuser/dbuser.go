// Package dbuser contains user related CRUD functionality.
package dbuser

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/business/sys/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Data manages the set of API's for user access.
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

// Create inserts a new user into the database.
func (d Data) Create(ctx context.Context, dbUsr DBUser) error {
	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, d.log, d.db, q, dbUsr); err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (d Data) Update(ctx context.Context, dbUsr DBUser) error {
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

	if err := database.NamedExecContext(ctx, d.log, d.db, q, dbUsr); err != nil {
		return fmt.Errorf("updating userID[%s]: %w", dbUsr.ID, err)
	}

	return nil
}

// Delete removes a user from the database.
func (d Data) Delete(ctx context.Context, userID string) error {
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

	if err := database.NamedExecContext(ctx, d.log, d.db, q, data); err != nil {
		return fmt.Errorf("deleting userID[%s]: %w", userID, err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (d Data) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]DBUser, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		users
	ORDER BY
		user_id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var dbUsrs []DBUser
	if err := database.NamedQuerySlice(ctx, d.log, d.db, q, data, &dbUsrs); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return dbUsrs, nil
}

// QueryByID gets the specified user from the database.
func (d Data) QueryByID(ctx context.Context, userID string) (DBUser, error) {
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

	var dbUsr DBUser
	if err := database.NamedQueryStruct(ctx, d.log, d.db, q, data, &dbUsr); err != nil {
		return DBUser{}, fmt.Errorf("selecting userID[%q]: %w", userID, err)
	}

	return dbUsr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (d Data) QueryByEmail(ctx context.Context, email string) (DBUser, error) {
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

	var dbUsr DBUser
	if err := database.NamedQueryStruct(ctx, d.log, d.db, q, data, &dbUsr); err != nil {
		return DBUser{}, fmt.Errorf("selecting email[%q]: %w", email, err)
	}

	return dbUsr, nil
}
