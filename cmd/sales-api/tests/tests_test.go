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

// Information about the users we have created for testing.
var adminAuthorization string
var adminID string
var userAuthorization string
var userID string

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

	kid := "4754d86b-7a6d-4df5-9c65-224741361492"
	kf := auth.NewSingleKeyFunc(kid, key.Public().(*rsa.PublicKey))
	authenticator, err := auth.NewAuthenticator(key, kid, "RS256", kf)
	if err != nil {
		panic(err)
	}

	a = handlers.API(test.Log, test.MasterDB, authenticator).(*web.App)

	// Create an admin user directly with our business logic. This creates an
	// initial user that we will use for admin validated endpoints.
	nu := user.NewUser{
		Email:           "admin@ardanlabs.com",
		Name:            "Admin User",
		Roles:           []string{auth.RoleAdmin, auth.RoleUser},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	admin, err := user.Create(tests.Context(), test.MasterDB, &nu, time.Now())
	if err != nil {
		panic(err)
	}
	adminID = admin.ID.Hex()

	tkn, err := user.Authenticate(tests.Context(), test.MasterDB, authenticator, time.Now(), nu.Email, nu.Password)
	if err != nil {
		panic(err)
	}

	adminAuthorization = "Bearer " + tkn.Token

	// Create a regular user to use when calling regular validated endpoints.
	nu = user.NewUser{
		Email:           "user@ardanlabs.com",
		Name:            "Regular User",
		Roles:           []string{auth.RoleUser},
		Password:        "concurrency",
		PasswordConfirm: "concurrency",
	}

	usr, err := user.Create(tests.Context(), test.MasterDB, &nu, time.Now())
	if err != nil {
		panic(err)
	}
	userID = usr.ID.Hex()

	tkn, err = user.Authenticate(tests.Context(), test.MasterDB, authenticator, time.Now(), nu.Email, nu.Password)
	if err != nil {
		panic(err)
	}

	userAuthorization = "Bearer " + tkn.Token

	return m.Run()
}
