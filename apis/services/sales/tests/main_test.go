package tests

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ardanlabs/service/apis/api/apptest"
	"github.com/ardanlabs/service/apis/api/authclient"
	authbuild "github.com/ardanlabs/service/apis/services/auth/build/all"
	authmux "github.com/ardanlabs/service/apis/services/auth/mux"
	salesbuild "github.com/ardanlabs/service/apis/services/sales/build/all"
	salesmux "github.com/ardanlabs/service/apis/services/sales/mux"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	var err error

	c, err = dbtest.StartDB()
	if err != nil {
		return 1, err
	}
	defer dbtest.StopDB(c)

	return m.Run(), nil
}

func startTest(t *testing.T, testName string) (*dbtest.Test, *apptest.AppTest) {
	dbTest := dbtest.NewTest(t, c, testName)

	// -------------------------------------------------------------------------

	authMux := authmux.WebAPI(authmux.Config{
		Shutdown: make(chan os.Signal, 1),
		Log:      dbTest.Log,
		Auth:     dbTest.Auth,
		DB:       dbTest.DB,
		BusDomain: authmux.BusDomain{
			Delegate: dbTest.BusDomain.Delegate,
			User:     dbTest.BusDomain.User,
		},
	}, authbuild.Routes())

	logFunc := func(ctx context.Context, msg string) {
		t.Logf("authapi: message: %s", msg)
	}

	server := httptest.NewServer(authMux)
	authClient := authclient.New(server.URL, logFunc)

	// -------------------------------------------------------------------------

	appTest := apptest.New(salesmux.WebAPI(salesmux.Config{
		Shutdown:   make(chan os.Signal, 1),
		Log:        dbTest.Log,
		AuthClient: authClient,
		DB:         dbTest.DB,
		BusDomain: salesmux.BusDomain{
			Delegate: dbTest.BusDomain.Delegate,
			Home:     dbTest.BusDomain.Home,
			Product:  dbTest.BusDomain.Product,
			User:     dbTest.BusDomain.User,
			VProduct: dbTest.BusDomain.VProduct,
		},
	}, salesbuild.Routes()))

	return dbTest, appTest
}
