// This program performs administrative tasks for the garage sale service.
//
// Run it with --cmd keygen or --cmd useradd

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/database"
	"github.com/ardanlabs/service/internal/platform/flag"
	"github.com/ardanlabs/service/internal/schema"
	"github.com/ardanlabs/service/internal/user"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func main() {

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "sales-admin : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration

	var cfg struct {
		CMD string `envconfig:"CMD"`
		DB  struct {
			User       string `default:"postgres"`
			Password   string `default:"postgres" json:"-"` // Prevent the marshalling of secrets.
			Host       string `default:"localhost"`
			Name       string `default:"postgres"`
			DisableTLS bool   `default:"false" split_words:"true"`
		}
		Auth struct {
			PrivateKeyFile string `default:"private.pem" envconfig:"PRIVATE_KEY_FILE"`
		}
		User struct {
			Email    string
			Password string
		}
	}

	if err := envconfig.Process("SALES", &cfg); err != nil {
		log.Fatalf("main : Parsing Config : %v", err)
	}

	if err := flag.Process(&cfg); err != nil {
		if err != flag.ErrHelp {
			log.Fatalf("main : Parsing Command Line : %v", err)
		}
		return // We displayed help.
	}

	// This is used for multiple commands below.
	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	var err error
	switch cfg.CMD {
	case "keygen":
		err = keygen(cfg.Auth.PrivateKeyFile)
	case "useradd":
		err = useradd(dbConfig, cfg.User.Email, cfg.User.Password)
	case "migrate":
		err = migrate(dbConfig)
	case "seed":
		err = seed(dbConfig)
	default:
		err = errors.New("Must provide --cmd")
	}

	if err != nil {
		log.Fatal(err)
	}
}

// keygen creates an x509 private key for signing auth tokens.
func keygen(path string) error {

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "generating keys")
	}

	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "creating private file")
	}

	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if err := pem.Encode(file, &block); err != nil {
		return errors.Wrap(err, "encoding to private file")
	}

	if err := file.Close(); err != nil {
		return errors.Wrap(err, "closing private file")
	}

	return nil
}

func useradd(cfg database.Config, email, pass string) error {

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if email == "" {
		return errors.New("Must provide --user_email")
	}
	if pass == "" {
		return errors.New("Must provide --user_password or set the env var SALES_USER_PASSWORD")
	}

	ctx := context.Background()

	nu := user.NewUser{
		Email:           email,
		Password:        pass,
		PasswordConfirm: pass,
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
	}

	usr, err := user.Create(ctx, db, &nu, time.Now())
	if err != nil {
		return err
	}

	fmt.Printf("User created with id: %v\n", usr.ID)
	return nil
}

func migrate(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	return schema.Migrate(db)
}

func seed(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	return schema.Seed(db)
}
