package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/internal/data"
	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/database"
	"github.com/pkg/errors"
)

// UserAdd adds new users into the database.
func UserAdd(cfg database.Config, email, password string) error {
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

	nu := data.NewUser{
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}
	u, err := data.Create.User(ctx, db, nu, time.Now())
	if err != nil {
		return errors.Wrap(err, "create user")
	}

	fmt.Println("user id:", u.ID)
	return nil
}
