// Package homedb contains home related CRUD functionality.
package homedb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/sqldb"
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

// NewWithTx constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) NewWithTx(tx sqldb.CommitRollbacker) (homebus.Storer, error) {
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

// Create inserts a new home into the database.
func (s *Store) Create(ctx context.Context, hme homebus.Home) error {
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

// Delete removes a home from the database.
func (s *Store) Delete(ctx context.Context, hme homebus.Home) error {
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

// Update replaces a home document in the database.
func (s *Store) Update(ctx context.Context, hme homebus.Home) error {
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

// Query retrieves a list of existing homes from the database.
func (s *Store) Query(ctx context.Context, filter homebus.QueryFilter, orderBy order.By, page page.Page) ([]homebus.Home, error) {
	data := map[string]any{
		"offset":        (page.Number() - 1) * page.RowsPerPage(),
		"rows_per_page": page.RowsPerPage(),
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

	var dbHmes []home
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbHmes); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	hmes, err := toBusHomes(dbHmes)
	if err != nil {
		return nil, err
	}

	return hmes, nil
}

// Count returns the total number of homes in the DB.
func (s *Store) Count(ctx context.Context, filter homebus.QueryFilter) (int, error) {
	data := map[string]any{}

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
		return 0, fmt.Errorf("db: %w", err)
	}

	return count.Count, nil
}

// QueryByID gets the specified home from the database.
func (s *Store) QueryByID(ctx context.Context, homeID uuid.UUID) (homebus.Home, error) {
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

	var dbHme home
	if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbHme); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return homebus.Home{}, fmt.Errorf("db: %w", homebus.ErrNotFound)
		}
		return homebus.Home{}, fmt.Errorf("db: %w", err)
	}

	return toBusHome(dbHme)
}

// QueryByUserID gets the specified home from the database by user id.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]homebus.Home, error) {
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

	var dbHmes []home
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbHmes); err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return toBusHomes(dbHmes)
}
