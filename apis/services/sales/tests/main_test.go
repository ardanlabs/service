package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	authbuild "github.com/ardanlabs/service/apis/services/auth/build/all"
	authmux "github.com/ardanlabs/service/apis/services/auth/mux"
	salesbuild "github.com/ardanlabs/service/apis/services/sales/build/all"
	salesmux "github.com/ardanlabs/service/apis/services/sales/mux"
	"github.com/ardanlabs/service/app/api/authclient"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
	"github.com/go-json-experiment/json"
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

func startTest(t *testing.T, testName string) (*dbtest.Test, *appTest) {
	dbTest := dbtest.NewTest(t, c, testName)

	// -------------------------------------------------------------------------

	authMux := authmux.WebAPI(authmux.Config{
		Log:  dbTest.Log,
		Auth: dbTest.Auth,
		DB:   dbTest.DB,
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

	appTest := newAppTest(salesmux.WebAPI(salesmux.Config{
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

// =============================================================================

type appTable struct {
	Name       string
	URL        string
	Token      string
	Method     string
	StatusCode int
	Model      any
	Resp       any
	ExpResp    any
	CmpFunc    func(got any, exp any) string
}

type appTest struct {
	handler http.Handler
}

func newAppTest(handler http.Handler) *appTest {
	return &appTest{
		handler: handler,
	}
}

func (at *appTest) run(t *testing.T, table []appTable, testName string) {
	for _, tt := range table {
		f := func(t *testing.T) {
			r := httptest.NewRequest(tt.Method, tt.URL, nil)
			w := httptest.NewRecorder()

			if tt.Model != nil {
				var b bytes.Buffer
				if err := json.MarshalWrite(&b, tt.Model, json.FormatNilSliceAsNull(true)); err != nil {
					t.Fatalf("Should be able to marshal the model : %s", err)
				}

				r = httptest.NewRequest(tt.Method, tt.URL, &b)
			}

			r.Header.Set("Authorization", "Bearer "+tt.Token)
			at.handler.ServeHTTP(w, r)

			if w.Code != tt.StatusCode {
				t.Fatalf("%s: Should receive a status code of %d for the response : %d", tt.Name, tt.StatusCode, w.Code)
			}

			if tt.StatusCode == http.StatusNoContent {
				return
			}

			if err := json.Unmarshal(w.Body.Bytes(), tt.Resp); err != nil {
				t.Fatalf("Should be able to unmarshal the response : %s", err)
			}

			diff := tt.CmpFunc(tt.Resp, tt.ExpResp)
			if diff != "" {
				t.Log("DIFF")
				t.Logf("%s", diff)
				t.Log("GOT")
				t.Logf("%#v", tt.Resp)
				t.Log("EXP")
				t.Logf("%#v", tt.ExpResp)
				t.Fatalf("Should get the expected response")
			}
		}

		t.Run(testName+"-"+tt.Name, f)
	}
}
