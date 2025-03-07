package apitest

import (
	"net/http/httptest"
	"testing"

	authbuild "github.com/ardanlabs/service/api/services/auth/build/all"
	salesbuild "github.com/ardanlabs/service/api/services/sales/build/all"
	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/app/sdk/authclient"
	"github.com/ardanlabs/service/app/sdk/mux"
	"github.com/ardanlabs/service/business/sdk/dbtest"
)

// New initialized the system to run a test.
func New(t *testing.T, testName string) *Test {
	db := dbtest.New(t, testName)

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
		Log: db.Log,
		DB:  db.DB,
		BusConfig: mux.BusConfig{
			UserBus: db.BusDomain.User,
		},
		AuthConfig: mux.AuthConfig{
			Auth: auth,
		},
	}, authbuild.Routes()))

	authClient := authclient.New(db.Log, server.URL)

	// -------------------------------------------------------------------------

	mux := mux.WebAPI(mux.Config{
		Log: db.Log,
		DB:  db.DB,
		BusConfig: mux.BusConfig{
			UserBus:     db.BusDomain.User,
			ProductBus:  db.BusDomain.Product,
			HomeBus:     db.BusDomain.Home,
			VProductBus: db.BusDomain.VProduct,
		},
		SalesConfig: mux.SalesConfig{
			AuthClient: authClient,
		},
	}, salesbuild.Routes())

	return &Test{
		DB:   db,
		Auth: auth,
		mux:  mux,
	}
}
