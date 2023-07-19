package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/data/dbmigrate"
	db "github.com/ardanlabs/service/business/data/dbsql/pgx"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

// Migrate creates the schema in the database.
func Migrate(cfg db.Config) error {
	db, err := db.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbmigrate.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")
	return nil
}
