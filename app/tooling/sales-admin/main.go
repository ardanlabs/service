// This program performs administrative tasks for the garage sale service.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ardanlabs/conf/v3"
	"github.com/ardanlabs/service/app/tooling/sales-admin/commands"
	database "github.com/ardanlabs/service/business/sys/database/pgx"
	"github.com/ardanlabs/service/foundation/vault"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var build = "develop"

type config struct {
	conf.Version
	Args conf.Args
	DB   struct {
		User         string `conf:"default:postgres"`
		Password     string `conf:"default:postgres,mask"`
		Host         string `conf:"default:database-service.sales-system.svc.cluster.local"`
		Name         string `conf:"default:postgres"`
		MaxIdleConns int    `conf:"default:2"`
		MaxOpenConns int    `conf:"default:0"`
		DisableTLS   bool   `conf:"default:true"`
	}
	Vault struct {
		KeysFolder string `conf:"default:zarf/keys/"`
		Address    string `conf:"default:http://vault-service.sales-system.svc.cluster.local:8200"`
		Token      string `conf:"default:mytoken,mask"`
		MountPath  string `conf:"default:secret"`
	}
}

func main() {
	if err := run(zap.NewNop().Sugar()); err != nil {
		if !errors.Is(err, commands.ErrHelp) {
			fmt.Println("ERROR", err)
		}
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	cfg := config{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "SALES"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		out, err := conf.String(&cfg)
		if err != nil {
			return fmt.Errorf("generating config for output: %w", err)
		}
		log.Infow("startup", "config", out)

		return fmt.Errorf("parsing config: %w", err)
	}

	return processCommands(cfg.Args, log, cfg)
}

// processCommands handles the execution of the commands specified on
// the command line.
func processCommands(args conf.Args, log *zap.SugaredLogger, cfg config) error {
	dbConfig := database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	}

	vaultConfig := vault.Config{
		Address:   cfg.Vault.Address,
		Token:     cfg.Vault.Token,
		MountPath: cfg.Vault.MountPath,
	}

	switch args.Num(0) {
	case "migrate":
		if err := commands.Migrate(dbConfig); err != nil {
			return fmt.Errorf("migrating database: %w", err)
		}

	case "seed":
		if err := commands.Seed(dbConfig); err != nil {
			return fmt.Errorf("seeding database: %w", err)
		}

	case "useradd":
		name := args.Num(1)
		email := args.Num(2)
		password := args.Num(3)
		if err := commands.UserAdd(log, dbConfig, name, email, password); err != nil {
			return fmt.Errorf("adding user: %w", err)
		}

	case "users":
		pageNumber := args.Num(1)
		rowsPerPage := args.Num(2)
		if err := commands.Users(log, dbConfig, pageNumber, rowsPerPage); err != nil {
			return fmt.Errorf("getting users: %w", err)
		}

	case "genkey":
		if err := commands.GenKey(); err != nil {
			return fmt.Errorf("key generation: %w", err)
		}

	case "gentoken":
		userID, err := uuid.Parse(args.Num(1))
		if err != nil {
			return fmt.Errorf("generating token: %w", err)
		}
		kid := args.Num(2)
		if err := commands.GenToken(log, dbConfig, vaultConfig, userID, kid); err != nil {
			return fmt.Errorf("generating token: %w", err)
		}

	case "vault":
		if err := commands.Vault(vaultConfig, cfg.Vault.KeysFolder); err != nil {
			return fmt.Errorf("setting private key: %w", err)
		}

	case "vault-init":
		if err := commands.VaultInit(vaultConfig); err != nil {
			return fmt.Errorf("initializing vault instance: %w", err)
		}

	default:
		fmt.Println("migrate:    create the schema in the database")
		fmt.Println("seed:       add data to the database")
		fmt.Println("useradd:    add a new user to the database")
		fmt.Println("users:      get a list of users from the database")
		fmt.Println("genkey:     generate a set of private/public key files")
		fmt.Println("gentoken:   generate a JWT for a user with claims")
		fmt.Println("vault:      load private keys into vault system")
		fmt.Println("vault-init: initialize a new vault instance")
		fmt.Println("provide a command to get more help.")
		return commands.ErrHelp
	}

	return nil
}
