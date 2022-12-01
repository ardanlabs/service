package userdb

import (
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/order"
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
	usr := user.User{
		ID:           dbUsr.ID,
		Name:         dbUsr.Name,
		Email:        dbUsr.Email,
		Roles:        dbUsr.Roles,
		PasswordHash: dbUsr.PasswordHash,
		Enabled:      dbUsr.Enabled,
		DateCreated:  dbUsr.DateCreated,
		DateUpdated:  dbUsr.DateUpdated,
	}
	usr.DateCreated = time.Date(usr.DateCreated.Year(), usr.DateCreated.Month(), usr.DateCreated.Day(), usr.DateCreated.Hour(), usr.DateCreated.Minute(), usr.DateCreated.Second(), usr.DateCreated.Nanosecond(), time.Local)
	usr.DateUpdated = time.Date(usr.DateUpdated.Year(), usr.DateUpdated.Month(), usr.DateUpdated.Day(), usr.DateUpdated.Hour(), usr.DateUpdated.Minute(), usr.DateUpdated.Second(), usr.DateUpdated.Nanosecond(), time.Local)

	return usr
}

func toCoreUserSlice(dbUsers []dbUser) []user.User {
	usrs := make([]user.User, len(dbUsers))
	for i, dbUsr := range dbUsers {
		usrs[i] = toCoreUser(dbUsr)
	}
	return usrs
}

// =============================================================================

// orderByfields is the map of fields that is used to translate between the
// application layer names and the database.
var orderByFields = map[string]string{
	user.OrderByID:      "user_id",
	user.OrderByName:    "name",
	user.OrderByEmail:   "email",
	user.OrderByRoles:   "roles",
	user.OrderByEnabled: "enabled",
}

// orderByClause validates the order by for correct fields and sql injection.
func orderByClause(orderBy order.By) (string, error) {
	if err := order.Validate(orderBy.Field, orderBy.Direction); err != nil {
		return "", err
	}

	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return by + " " + orderBy.Direction, nil
}
