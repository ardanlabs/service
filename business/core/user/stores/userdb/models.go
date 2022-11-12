package userdb

import (
	"time"
	"unsafe"

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

	// We can't use dbUser(usr) because pq requires us to use their
	// StringArray type so we are performing a casting operation.
	return *(*dbUser)(unsafe.Pointer(&usr))
}

func toCoreUser(dbUsr dbUser) user.User {

	// We can't use user.User(dbUsr) because pq requires us to use their
	// StringArray type so we are performing a casting operation.
	return *(*user.User)(unsafe.Pointer(&dbUsr))
}

func toCoreUserSlice(dbUsers []dbUser) []user.User {
	usrs := make([]user.User, len(dbUsers))
	for i, dbUsr := range dbUsers {
		usrs[i] = toCoreUser(dbUsr)
	}
	return usrs
}
