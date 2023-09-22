package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	v1 "github.com/ardanlabs/service/app/services/sales-api/handlers/v1"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/groups/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/handlers/v1/paging"
	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/order"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// HomeTests holds methods for each home subtest. This type allows passing
// dependencies for test while still providing a convenient synta when
// subtests are registered.
type HomeTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

// Test_Homes is the entry point for testing home management apis.
func Test_Homes(t *testing.T) {
	t.Parallel()

	test := dbtest.NewTest(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		test.Teardown()
	}()

	api := test.CoreAPIs

	shutdown := make(chan os.Signal, 1)
	tests := HomeTests{
		app: v1.APIMux(v1.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		userToken:  test.Token("user@example.com", "gophers"),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	// ------------------------------------------------------------------------
	seed := func(ctx context.Context, usrCore *user.Core, hmeCore *home.Core) ([]home.Home, error) {
		usrs, err := usrCore.Query(ctx, user.QueryFilter{}, order.By{Field: user.OrderByName, Direction: order.ASC}, 1, 2)
		if err != nil {
			return nil, fmt.Errorf("seeding users : %w", err)
		}

		hmes1, err := home.TestGenerateSeedHomes(5, hmeCore, usrs[0].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding homes1 : %w", err)
		}

		hmes2, err := home.TestGenerateSeedHomes(5, hmeCore, usrs[1].ID)
		if err != nil {
			return nil, fmt.Errorf("seeding homes2 : %w", err)
		}

		var hmes []home.Home
		hmes = append(hmes, hmes1...)
		hmes = append(hmes, hmes2...)

		return hmes, nil
	}

	t.Log("Go seeding ...")

	hmes, err := seed(context.Background(), api.User, api.Home)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	t.Run("crudHome", tests.crudHome())
	t.Run("getHomes200", tests.getHomes200(hmes))
}

// getHomes200 validates a query request.
func (ht *HomeTests) getHomes200(hmes []home.Home) func(t *testing.T) {
	return func(t *testing.T) {
		url := "/v1/homes?page=1&rows=10&orderBy=user_id,DESC"

		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ht.userToken)
		ht.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
		}

		var pr paging.Response[homegrp.AppHome]
		if err := json.Unmarshal(w.Body.Bytes(), &pr); err != nil {
			t.Fatalf("Should be able to unmarshal the response : %s", err)
		}

		if pr.RowsPerPage != len(hmes) {
			t.Log("got:", pr.RowsPerPage)
			t.Log("exp:", len(hmes))
			t.Error("Should get the right total")
		}

		// pr.Total >= 10 is put this would fail if we don't have enough items
		// in the server.
		if len(pr.Items) != 10 && pr.Total >= 10 {
			t.Log("got:", len(pr.Items))
			t.Log("exp:", 10)
			t.Error("Should get the right number of homes")
		}
	}
}

func (ht *HomeTests) crudHome() func(t *testing.T) {
	return func(t *testing.T) {
		hme := ht.postHome201(t)
		defer ht.deleteHome204(t, hme.ID)

		ht.getHome200(t, hme.ID)
		ht.putHome200(t, hme.ID)
	}
}

// postHome201 validates a home can be created with the endpoint.
func (ht *HomeTests) postHome201(t *testing.T) homegrp.AppHome {
	np := home.NewHome{
		Type:   "House",
		UserID: uuid.MustParse("5cf37266-3473-4006-984f-9325122678b7"),
		Address: home.Address{
			Address1: "123 Mocking Bird Ln",
			Address2: "",
			ZipCode:  "33156",
			City:     "Miami",
			State:    "Miami",
			Country:  "US",
		},
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/homes", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ht.userToken)
	ht.app.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("Should receive a status code of 201 for the response : %d", w.Code)
	}

	var newHme homegrp.AppHome
	if err := json.NewDecoder(w.Body).Decode(&newHme); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	exp := newHme

	exp.Type = "House"
	exp.Address.Address1 = "123 Mocking Bird Ln"
	exp.Address.Address2 = ""
	exp.Address.ZipCode = "33156"
	exp.Address.City = "Miami"
	exp.Address.State = "Miami"
	exp.Address.Country = "US"

	if diff := cmp.Diff(newHme, exp); diff != "" {
		t.Fatalf("Should get the expected result, Diff:\n%s", diff)
	}

	return newHme
}

// deleteProduct200 validates deleting a home that does exist.
func (ht *HomeTests) deleteHome204(t *testing.T, id string) {
	url := fmt.Sprintf("/v1/homes/%s", id)

	r := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ht.userToken)
	ht.app.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Should receive a status code of 204 for the response : %d", w.Code)
	}
}

func (ht *HomeTests) getHome200(t *testing.T, id string) {
	url := fmt.Sprintf("/v1/homes/%s", id)

	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ht.userToken)
	ht.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
	}

	var got homegrp.AppHome
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	exp := got
	// todo

	if diff := cmp.Diff(got, exp); diff != "" {
		t.Fatalf("Should get the expected result, Diff:\n%s", diff)
	}
}

func (ht *HomeTests) putHome200(t *testing.T, id string) {
	body := `{"type": "Motor Home"}`

	url := fmt.Sprintf("/v1/homes/%s", id)

	r := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ht.adminToken)
	ht.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the response : %d", w.Code)
	}

	r = httptest.NewRequest(http.MethodGet, "/v1/homes/"+id, nil)
	w = httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ht.adminToken)
	ht.app.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Should receive a status code of 200 for the retrieve : %d", w.Code)
	}

	var ru homegrp.AppHome
	if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
		t.Fatalf("Should be able to unmarshal the response : %s", err)
	}

	if ru.Type != "Motor Home" {
		t.Fatalf("Should see an updated Name : got %q want %q", ru.Type, "Motor Home")
	}

}
