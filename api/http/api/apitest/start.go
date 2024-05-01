package apitest

import (
	"net/http/httptest"
	"testing"

	authbuild "github.com/ardanlabs/service/api/cmd/services/auth/build/all"
	salesbuild "github.com/ardanlabs/service/api/cmd/services/sales/build/all"
	"github.com/ardanlabs/service/api/http/api/mux"
	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
)

// StartTest initialized the system to run a test.
func StartTest(t *testing.T, c *docker.Container, testName string) *Test {
	db := dbtest.NewDatabase(t, c, testName)

	// -------------------------------------------------------------------------

	auth, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &KeyStore{},
	})
	if err != nil {
		t.Fatal(err)
	}

	// -------------------------------------------------------------------------

	server := httptest.NewServer(mux.WebAPI(mux.Config{
		Log:  db.Log,
		Auth: auth,
		DB:   db.DB,
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, server.URL)

	// -------------------------------------------------------------------------

	mux := mux.WebAPI(mux.Config{
		Log:        db.Log,
		AuthClient: authClient,
		DB:         db.DB,
	}, salesbuild.Routes())

	return New(db, auth, mux)
}
