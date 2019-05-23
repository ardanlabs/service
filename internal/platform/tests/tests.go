package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/database/databasetest"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/schema"
	"github.com/ardanlabs/service/internal/user"
	"github.com/jmoiron/sqlx"
	"github.com/pborman/uuid"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// These are the IDs in the seed data for admin@example.com and
// user@example.com.
const (
	AdminID = "5cf37266-3473-4006-984f-9325122678b7"
	UserID  = "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
)

// Test owns state for running and shutting down tests.
type Test struct {
	DB            *sqlx.DB
	Log           *log.Logger
	Authenticator *auth.Authenticator

	cleanup func()
}

// New creates a database, seeds it, constructs an authenticator.
func New(t *testing.T) *Test {
	t.Helper()

	var test Test

	// Initialize and seed database. Store the cleanup function call later.
	test.DB, test.cleanup = databasetest.Setup(t)

	if err := schema.Seed(test.DB); err != nil {
		t.Fatal(err)
	}

	// Create the logger to use.
	test.Log = log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Create RSA keys to enable authentication in our service.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Build an authenticator using this static key.
	kid := "4754d86b-7a6d-4df5-9c65-224741361492"
	kf := auth.NewSimpleKeyLookupFunc(kid, key.Public().(*rsa.PublicKey))
	test.Authenticator, err = auth.NewAuthenticator(key, kid, "RS256", kf)
	if err != nil {
		t.Fatal(err)
	}

	return &test
}

// Teardown releases any resources used for the test.
func (test *Test) Teardown() {
	test.cleanup()
}

// Token generates an auhenticated token for a user.
func (test *Test) Token(t *testing.T, email, pass string) string {
	t.Helper()

	claims, err := user.Authenticate(
		context.Background(), test.DB, time.Now(),
		email, pass,
	)
	if err != nil {
		t.Fatal(err)
	}

	tkn, err := test.Authenticator.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	return tkn
}

// Context returns an app level context for testing.
func Context() context.Context {
	values := web.Values{
		TraceID: uuid.New(),
		Now:     time.Now(),
	}

	return context.WithValue(context.Background(), web.KeyValues, &values)
}
