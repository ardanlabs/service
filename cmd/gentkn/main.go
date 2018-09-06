package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

func main() {

	// Define and process configuration for entire tool.
	var cfg struct {
		KeyID string `envconfig:"KEY_ID" default:"default"`

		Claims struct {
			Audience string   `envconfig:"AUDIENCE"`
			Subject  string   `envconfig:"SUBJECT"`
			Roles    []string `envconfig:"ROLES"`
		} `envconfig:"CLAIMS"`
	}
	if err := envconfig.Process("GENTKN", &cfg); err != nil {
		log.Fatalf("main : Parsing Config : %v", err)
	}

	/* TODO: Fix non-flag processing in flag pkg.
	if err := flag.Process(&cfg); err != nil {
		if err != flag.ErrHelp {
			log.Fatalf("main : Parsing Command Line : %v", err)
		}
		return // We displayed help.
	}
	*/

	// Source the private key.
	var keyRdr io.Reader
	if len(os.Args) < 2 {
		log.Fatal("main : Missing Required Param <key-file>")
	}
	switch file := os.Args[1]; file {
	case "-":
		keyRdr = os.Stdin
	default:
		f, err := os.Open(file)
		if err != nil {
			log.Fatalf("main : Opening Key File : %v", err)
		}
		defer f.Close()
		keyRdr = f
	}
	keyContents, err := ioutil.ReadAll(keyRdr)
	if err != nil {
		log.Fatalf("main : Reading Key : %v", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyContents)
	if err != nil {
		log.Fatalf("main : Parsing RSA Private Key : %v", err)
	}

	// Build the claims.
	now := time.Now()
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Audience:  cfg.Claims.Audience,
			ExpiresAt: now.Add(7 * 24 * time.Hour).Unix(),
			Id:        uuid.New().String(),
			IssuedAt:  now.Unix(),
			Subject:   cfg.Claims.Subject,
		},
		Roles: cfg.Claims.Roles,
	}
	if err := claims.Valid(); err != nil {
		log.Fatalf("main : Invalid Claims : %v", err)
	}

	// Generate and sign the token.
	tkn, err := auth.GenerateToken(cfg.KeyID, key, jwt.SigningMethodRS256, claims)
	if err != nil {
		log.Fatalf("main : Generating Token : %v", err)
	}
	fmt.Println(tkn)
}
