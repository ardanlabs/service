package commands

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ardanlabs/service/business/auth"
	"github.com/ardanlabs/service/business/data/user"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/pkg/errors"
)

// UserAdd adds new users into the database.
func UserAdd(traceID string, log *log.Logger, cfg database.Config, email, password string) error {
	if email == "" || password == "" {
		fmt.Println("help: useradd <email> <password>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u := user.New(log, db)

	nu := user.NewUser{
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}

	usr, err := u.Create(ctx, traceID, nu, time.Now())
	if err != nil {
		return errors.Wrap(err, "create user")
	}

	fmt.Println("user id:", usr.ID)
	return nil
}
