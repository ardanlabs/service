package commands

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/service/business/data/user"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/pkg/errors"
)

// Users retrieves all users from the database.
func Users(traceID string, log *log.Logger, cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return errors.Wrap(err, "connect database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u := user.New(log, db)

	users, err := u.Query(ctx, traceID)
	if err != nil {
		return errors.Wrap(err, "retrieve users")
	}

	return json.NewEncoder(os.Stdout).Encode(users)
}
