package tests

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/ardanlabs/service/cmd/sales-api/handlers"
	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/ardanlabs/service/internal/user"
)

var a *web.App
var test *tests.Test
var adminAuthorization string
var userAuthorization string

// TestMain is the entry point for testing.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	test = tests.New()
	defer test.TearDown()

	// Create RSA keys to enable authentication in our service.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	userAuth := handlers.UserAuth{
		Key:   key,
		KeyID: "4754d86b-7a6d-4df5-9c65-224741361492",
		Alg:   "RS256",
	}

	a = handlers.API(test.Log, test.MasterDB, userAuth).(*web.App)

	// Create an admin user directly with our business logic. This creates an
	// initial user that we will use for admin validated endpoints.
	admin := user.NewUser{
		Email:           "admin@ardanlabs.com",
		Name:            "Admin User",
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	if _, err := user.Create(tests.Context(), test.MasterDB, &admin, time.Now()); err != nil {
		panic(err)
	}

	tkn, err := user.Authenticate(tests.Context(), test.MasterDB, time.Now(), userAuth.Key, userAuth.KeyID, userAuth.Alg, admin.Email, admin.Password)
	if err != nil {
		panic(err)
	}

	adminAuthorization = "Bearer " + tkn.Token

	// Create a regular user to use when calling regular validated endpoints.
	u := user.NewUser{
		Email:           "user@ardanlabs.com",
		Name:            "Regular User",
		Roles:           []string{auth.RoleUser},
		Password:        "concurrency",
		PasswordConfirm: "concurrency",
	}

	if _, err := user.Create(tests.Context(), test.MasterDB, &u, time.Now()); err != nil {
		panic(err)
	}

	tkn, err = user.Authenticate(tests.Context(), test.MasterDB, time.Now(), userAuth.Key, userAuth.KeyID, userAuth.Alg, u.Email, u.Password)
	if err != nil {
		panic(err)
	}

	userAuthorization = "Bearer " + tkn.Token

	return m.Run()
}
