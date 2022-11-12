package userdb

import (
	"time"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/lib/pq"
)

// dbUser represent the structure we need for moving data
// between the app and the database.
type dbUser struct {
	ID           string         `db:"user_id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	Roles        pq.StringArray `db:"roles"`
	PasswordHash []byte         `db:"password_hash"`
	Enabled      bool           `db:"enabled"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  time.Time      `db:"date_updated"`
}

// =============================================================================

func toDBUser(usr user.User) dbUser {
	return dbUser{
		ID:           usr.ID,
		Name:         usr.Name,
		Email:        usr.Email,
		Roles:        usr.Roles,
		PasswordHash: usr.PasswordHash,
		Enabled:      usr.Enabled,
		DateCreated:  usr.DateCreated,
		DateUpdated:  usr.DateUpdated,
	}
}

func toCoreUser(dbUsr dbUser) user.User {
	return user.User{
		ID:           dbUsr.ID,
		Name:         dbUsr.Name,
		Email:        dbUsr.Email,
		Roles:        dbUsr.Roles,
		PasswordHash: dbUsr.PasswordHash,
		Enabled:      dbUsr.Enabled,
		DateCreated:  dbUsr.DateCreated,
		DateUpdated:  dbUsr.DateUpdated,
	}
}

func toCoreUserSlice(dbUsers []dbUser) []user.User {
	usrs := make([]user.User, len(dbUsers))
	for i, dbUsr := range dbUsers {
		usrs[i] = toCoreUser(dbUsr)
	}
	return usrs
}
