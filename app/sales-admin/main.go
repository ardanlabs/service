// This program performs administrative tasks for the garage sale service.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/ardanlabs/service/app/sales-admin/commands"
	"github.com/ardanlabs/service/foundation/database"
	"github.com/pkg/errors"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	if err := run(); err != nil {
		if errors.Cause(err) != commands.ErrHelp {
			log.Printf("error: %s", err)
		}
		os.Exit(1)
	}
}

func run() error {

	// =========================================================================
	// Configuration

	var cfg struct {
		conf.Version
		Args conf.Args
		DB   struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:0.0.0.0"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:false"`
		}
	}
	cfg.Version.SVN = build
	cfg.Version.Desc = "copyright information here"

	args := os.Args[1:]
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	const prefix = "SALES"
	if err := conf.Parse(args, prefix, &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config version")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
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

	var command string
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "migrate":
		if err := commands.Migrate(dbConfig); err != nil {
			return errors.Wrap(err, "migrating database")
		}

	case "seed":
		if err := commands.Seed(dbConfig); err != nil {
			return errors.Wrap(err, "seeding database")
		}

	case "useradd":
		email := cfg.Args.Num(1)
		password := cfg.Args.Num(2)
		if err := commands.UserAdd(dbConfig, email, password); err != nil {
			return errors.Wrap(err, "adding user")
		}

	case "users":
		if err := commands.Users(dbConfig); err != nil {
			return errors.Wrap(err, "getting users")
		}

	case "keygen":
		if err := commands.KeyGen(); err != nil {
			return errors.Wrap(err, "key generation")
		}

	case "gentoken":
		email := cfg.Args.Num(1)
		privateKeyFile := cfg.Args.Num(2)
		algorithm := cfg.Args.Num(3)
		if err := commands.GenToken(dbConfig, email, privateKeyFile, algorithm); err != nil {
			return errors.Wrap(err, "generating token")
		}

	default:
		fmt.Println("migrate: create the schema in the database")
		fmt.Println("seed: add data to the database")
		fmt.Println("useradd: add a new user to the database")
		fmt.Println("users: get a list of users from the database")
		fmt.Println("keygen: generate a set of private/public key files")
		fmt.Println("gentoken: generate a JWT for a user with claims")
		fmt.Println("provide a command to get more help.")
		return commands.ErrHelp
	}

	return nil
}
