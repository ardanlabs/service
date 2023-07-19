package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/data/dbmigrate"
	db "github.com/ardanlabs/service/business/data/dbsql/pgx"
)

// Seed loads test data into the database.
func Seed(cfg db.Config) error {
	db, err := db.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbmigrate.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	fmt.Println("seed data complete")
	return nil
}
