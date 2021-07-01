// This program performs administrative tasks for the garage sale service.

package main

import (
	"fmt"
	"os"

	"github.com/ardanlabs/service/app/sales-admin/commands"
	"github.com/ardanlabs/service/business/config"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log := logger.New("ADMIN")
	defer log.Sync()

	if err := run(log); err != nil {
		if errors.Cause(err) != commands.ErrHelp {
			log.Errorw("", zap.Error(err))
		}
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// =========================================================================
	// Configuration
	cfg, err := config.Configuration(log, build)
	if err != nil {
		return err
	}

	// =========================================================================
	// Commands

	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	traceID := "00000000-0000-0000-0000-000000000000"

	switch cfg.Args.Num(0) {
	case "migrate":
		if err := commands.Migrate(dbConfig); err != nil {
			return errors.Wrap(err, "migrating database")
		}

	case "seed":
		if err := commands.Seed(dbConfig); err != nil {
			return errors.Wrap(err, "seeding database")
		}

	case "useradd":
		name := cfg.Args.Num(1)
		email := cfg.Args.Num(2)
		password := cfg.Args.Num(3)
		if err := commands.UserAdd(traceID, log, dbConfig, name, email, password); err != nil {
			return errors.Wrap(err, "adding user")
		}

	case "users":
		pageNumber := cfg.Args.Num(1)
		rowsPerPage := cfg.Args.Num(2)
		if err := commands.Users(traceID, log, dbConfig, pageNumber, rowsPerPage); err != nil {
			return errors.Wrap(err, "getting users")
		}

	case "genkey":
		if err := commands.GenKey(); err != nil {
			return errors.Wrap(err, "key generation")
		}

	case "gentoken":
		id := cfg.Args.Num(1)
		privateKeyFile := cfg.Args.Num(2)
		algorithm := cfg.Args.Num(3)
		if err := commands.GenToken(traceID, log, dbConfig, id, privateKeyFile, algorithm); err != nil {
			return errors.Wrap(err, "generating token")
		}

	default:
		fmt.Println("migrate: create the schema in the database")
		fmt.Println("seed: add data to the database")
		fmt.Println("useradd: add a new user to the database")
		fmt.Println("users: get a list of users from the database")
		fmt.Println("genkey: generate a set of private/public key files")
		fmt.Println("gentoken: generate a JWT for a user with claims")
		fmt.Println("provide a command to get more help.")
		return commands.ErrHelp
	}

	return nil
}
