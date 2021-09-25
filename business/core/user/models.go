package user

import (
	"time"
	"unsafe"

	"github.com/ardanlabs/service/business/data/dbuser"
)

// User represents an individual user.
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Roles        []string  `json:"roles"`
	PasswordHash []byte    `json:"-"`
	DateCreated  time.Time `json:"date_created"`
	DateUpdated  time.Time `json:"date_updated"`
}

// =============================================================================

func toUser(dbUsr dbuser.DBUser) User {
	pu := (*User)(unsafe.Pointer(&dbUsr))
	return *pu
}

func toUserSlice(dbUsrs []dbuser.DBUser) []User {
	users := make([]User, len(dbUsrs))
	for i, dbUsr := range dbUsrs {
		users[i] = toUser(dbUsr)
	}
	return users
}
