// Package homedb contains home related CRUD functionality.
package homedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/data/sqldb"
	"github.com/ardanlabs/service/business/data/transaction"
	"github.com/ardanlabs/service/business/web/v1/order"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for home database access.
type Store struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// ExecuteUnderTransaction constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) ExecuteUnderTransaction(tx transaction.Transaction) (home.Storer, error) {
	ec, err := sqldb.GetExtContext(tx)
	if err != nil {
		return nil, err
	}

	store := Store{
		log: s.log,
		db:  ec,
	}

	return &store, nil
}

// Create adds a Home to the sqldb. It returns an error if something went wrong
func (s *Store) Create(ctx context.Context, hme home.Home) error {
	const q = `
    INSERT INTO homes
        (home_id, user_id, type, address_1, address_2, zip_code, city, state, country, date_created, date_updated)
    VALUES
        (:home_id, :user_id, :type, :address_1, :address_2, :zip_code, :city, :state, :country, :date_created, :date_updated)`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBHome(hme)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes the Home identified by a given ID.
func (s *Store) Delete(ctx context.Context, hme home.Home) error {
	data := struct {
		ID string `db:"home_id"`
	}{
		ID: hme.ID.String(),
	}

	const q = `
    DELETE FROM
	    homes
	WHERE
	  	home_id = :home_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing homes from the database.
func (s *Store) Query(ctx context.Context, filter home.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]home.Home, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `
    SELECT
	    home_id, user_id, type, address_1, address_2, zip_code, city, state, country, date_created, date_updated
	FROM
	  	homes`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var dbHmes []dbHome
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbHmes); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	hmes, err := toCoreHomeSlice(dbHmes)
	if err != nil {
		return nil, err
	}

	return hmes, nil
}

// Update modifies data about a Home. It will error if the specified ID is
// invalid or does not reference an existing Home.
func (s *Store) Update(ctx context.Context, hme home.Home) error {
	const q = `
    UPDATE
        homes
    SET
        "address_1"     = :address_1,
        "address_2"     = :address_2,
        "zip_code"      = :zip_code,
        "city"          = :city,
        "state"         = :state,
        "country"       = :country,
        "type"          = :type,
        "date_updated"  = :date_updated
    WHERE
        home_id = :home_id`

	if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBHome(hme)); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Count returns the total number of homes in the DB.
func (s *Store) Count(ctx context.Context, filter home.QueryFilter) (int, error) {
	data := map[string]interface{}{}

	const q = `
    SELECT
        count(1)
    FROM
        homes`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID finds the home identified by a given ID.
func (s *Store) QueryByID(ctx context.Context, homeID uuid.UUID) (home.Home, error) {
	data := struct {
		ID string `db:"home_id"`
	}{
		ID: homeID.String(),
	}

	const q = `
    SELECT
	  	home_id, user_id, type, address_1, address_2, zip_code, city, state, country, date_created, date_updated
    FROM
        homes
    WHERE
        home_id = :home_id`

	var dbHme dbHome
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbHme); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return home.Home{}, fmt.Errorf("namedquerystruct: %w", home.ErrNotFound)
		}
		return home.Home{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreHome(dbHme)
}

// QueryByUserID finds the home identified by a given User ID.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]home.Home, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
	    home_id, user_id, type, address_1, address_2, zip_code, city, state, country, date_created, date_updated
	FROM
		homes
	WHERE
		user_id = :user_id`

	var dbHmes []dbHome
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbHmes); err != nil {
		return nil, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreHomeSlice(dbHmes)
}
