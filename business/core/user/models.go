package user

import (
	"time"
	"unsafe"

	"github.com/ardanlabs/service/business/data/store/user"
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

func toUser(dbUsr user.DBUser) User {
	pu := (*User)(unsafe.Pointer(&dbUsr))
	return *pu
}

func toUserSlice(dbUsrs []user.DBUser) []User {
	users := make([]User, len(dbUsrs))
	for i, dbUsr := range dbUsrs {
		users[i] = toUser(dbUsr)
	}
	return users
}
