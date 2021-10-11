// Package dbtest contains supporting code for running tests that hit the DB.
package dbtest

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	dbUser "github.com/ardanlabs/service/business/core/user/db"
	"github.com/ardanlabs/service/business/data/dbschema"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/database"
	"github.com/ardanlabs/service/foundation/docker"
	"github.com/ardanlabs/service/foundation/keystore"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// DBContainer provides configuration for a container to run.
type DBContainer struct {
	Image string
	Port  string
	Args  []string
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty. It returns
// the database to use as well as a function to call at the end of the test.
func NewUnit(t *testing.T, dbc DBContainer) (*zap.SugaredLogger, *sqlx.DB, func()) {
	c := docker.StartContainer(t, dbc.Image, dbc.Port, dbc.Args...)

	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Log("Waiting for database to be ready ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbschema.Migrate(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("Migrating error: %s", err)
	}

	if err := dbschema.Seed(ctx, db); err != nil {
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("Seeding error: %s", err)
	}

	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&buf)
	log := zap.New(
		zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel)).
		Sugar()

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()
		db.Close()
		docker.StopContainer(t, c.ID)

		log.Sync()

		writer.Flush()
		fmt.Println("******************** LOGS ********************")
		fmt.Print(buf.String())
		fmt.Println("******************** LOGS ********************")
	}

	return log, db, teardown
}

// Test owns state for running and shutting down tests.
type Test struct {
	DB       *sqlx.DB
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	Teardown func()

	t *testing.T
}

// NewIntegration creates a database, seeds it, constructs an authenticator.
func NewIntegration(t *testing.T, dbc DBContainer) *Test {
	log, db, teardown := NewUnit(t, dbc)

	// Create RSA keys to enable authentication in our service.
	keyID := "4754d86b-7a6d-4df5-9c65-224741361492"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Build an authenticator using this private key and id for the key store.
	auth, err := auth.New(keyID, keystore.NewMap(map[string]*rsa.PrivateKey{keyID: privateKey}))
	if err != nil {
		t.Fatal(err)
	}

	test := Test{
		DB:       db,
		Log:      log,
		Auth:     auth,
		t:        t,
		Teardown: teardown,
	}

	return &test
}

// Token generates an authenticated token for a user.
func (test *Test) Token(email, pass string) string {
	test.t.Log("Generating token for test ...")

	store := dbUser.NewStore(test.Log, test.DB)
	dbUsr, err := store.QueryByEmail(context.Background(), email)
	if err != nil {
		return ""
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   dbUsr.ID,
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: dbUsr.Roles,
	}

	token, err := test.Auth.GenerateToken(claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return token
}

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}
