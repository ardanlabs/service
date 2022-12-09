package userdb

import (
	"fmt"
	"net/mail"
	"time"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/google/uuid"
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
		ID:           usr.ID.String(),
		Name:         usr.Name,
		Email:        usr.Email.Address,
		Roles:        usr.Roles,
		PasswordHash: usr.PasswordHash,
		Enabled:      usr.Enabled,
		DateCreated:  usr.DateCreated.UTC(),
		DateUpdated:  usr.DateUpdated.UTC(),
	}
}

func toCoreUser(dbUsr dbUser) user.User {
	addr, err := mail.ParseAddress(dbUsr.Email)
	if err != nil {
		addr.Name = "unknown"
		addr.Address = dbUsr.Email
	}

	usr := user.User{
		ID:           uuid.MustParse(dbUsr.ID),
		Name:         dbUsr.Name,
		Email:        *addr,
		Roles:        dbUsr.Roles,
		PasswordHash: dbUsr.PasswordHash,
		Enabled:      dbUsr.Enabled,
		DateCreated:  dbUsr.DateCreated.In(time.Local),
		DateUpdated:  dbUsr.DateUpdated.In(time.Local),
	}

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
