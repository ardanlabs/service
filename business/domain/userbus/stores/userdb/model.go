package userdb

import (
	"database/sql"
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/sqldb/dbarray"
	"github.com/google/uuid"
)

type user struct {
	ID           uuid.UUID      `db:"user_id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	Roles        dbarray.String `db:"roles"`
	PasswordHash []byte         `db:"password_hash"`
	Department   sql.NullString `db:"department"`
	Enabled      bool           `db:"enabled"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  time.Time      `db:"date_updated"`
}

func toDBUser(bus userbus.User) user {
	roles := make([]string, len(bus.Roles))
	for i, role := range bus.Roles {
		roles[i] = role.String()
	}

	return user{
		ID:           bus.ID,
		Name:         bus.Name.String(),
		Email:        bus.Email.Address,
		Roles:        roles,
		PasswordHash: bus.PasswordHash,
		Department: sql.NullString{
			String: bus.Department,
			Valid:  bus.Department != "",
		},
		Enabled:     bus.Enabled,
		DateCreated: bus.DateCreated.UTC(),
		DateUpdated: bus.DateUpdated.UTC(),
	}
}

func toBusUser(db user) (userbus.User, error) {
	addr := mail.Address{
		Address: db.Email,
	}

	roles := make([]userbus.Role, len(db.Roles))
	for i, value := range db.Roles {
		var err error
		roles[i], err = userbus.Roles.Parse(value)
		if err != nil {
			return userbus.User{}, fmt.Errorf("parse role: %w", err)
		}
	}

	name, err := userbus.Names.Parse(db.Name)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse name: %w", err)
	}

	bus := userbus.User{
		ID:           db.ID,
		Name:         name,
		Email:        addr,
		Roles:        roles,
		PasswordHash: db.PasswordHash,
		Enabled:      db.Enabled,
		Department:   db.Department.String,
		DateCreated:  db.DateCreated.In(time.Local),
		DateUpdated:  db.DateUpdated.In(time.Local),
	}

	return bus, nil
}

func toBusUsers(dbs []user) ([]userbus.User, error) {
	bus := make([]userbus.User, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toBusUser(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
