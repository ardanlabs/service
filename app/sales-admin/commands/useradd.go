package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/data/user"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// UserAdd adds new users into the database.
func UserAdd(traceID string, log *zap.SugaredLogger, cfg database.Config, name, email, password string) error {
	if name == "" || email == "" || password == "" {
		fmt.Println("help: useradd <name> <email> <password>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := user.NewStore(log, db)

	nu := user.NewUser{
		Name:            name,
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}

	usr, err := store.Create(ctx, traceID, nu, time.Now())
	if err != nil {
		return errors.Wrap(err, "create user")
	}

	fmt.Println("user id:", usr.ID)
	return nil
}
