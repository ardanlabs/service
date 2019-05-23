package user_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/user"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

var test *tests.Test

// TestMain is the entry point for testing.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	test = tests.New()
	defer test.TearDown()
	return m.Run()
}

// TestUser validates the full set of CRUD operations on User values.
func TestUser(t *testing.T) {
	defer tests.Recover(t)

	t.Log("Given the need to work with User records.")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()
			dbConn := test.MasterDB.Copy()
			defer dbConn.Close()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			// claims is information about the person making the request.
			claims := auth.NewClaims(bson.NewObjectId().Hex(), []string{auth.RoleAdmin}, now, time.Hour)

			nu := user.NewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			u, err := user.Create(ctx, dbConn, &nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			savedU, err := user.Retrieve(ctx, claims, dbConn, u.ID.Hex())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user by ID: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user by ID.", tests.Success)

			if diff := cmp.Diff(u, savedU); diff != "" {
				t.Fatalf("\t%s\tShould get back the same user. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same user.", tests.Success)

			upd := user.UpdateUser{
				Name:  tests.StringPointer("Jacob Walker"),
				Email: tests.StringPointer("jacob@ardanlabs.com"),
			}

			if err := user.Update(ctx, dbConn, u.ID.Hex(), &upd, now); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", tests.Success)

			savedU, err = user.Retrieve(ctx, claims, dbConn, u.ID.Hex())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", tests.Success)

			if savedU.Name != *upd.Name {
				t.Errorf("\t%s\tShould be able to see updates to Name.", tests.Failed)
				t.Log("\t\tGot:", savedU.Name)
				t.Log("\t\tExp:", *upd.Name)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Name.", tests.Success)
			}

			if savedU.Email != *upd.Email {
				t.Errorf("\t%s\tShould be able to see updates to Email.", tests.Failed)
				t.Log("\t\tGot:", savedU.Email)
				t.Log("\t\tExp:", *upd.Email)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Email.", tests.Success)
			}

			if err := user.Delete(ctx, dbConn, u.ID.Hex()); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)

			savedU, err = user.Retrieve(ctx, claims, dbConn, u.ID.Hex())
			if errors.Cause(err) != user.ErrNotFound {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", tests.Success)
		}
	}
}

// mockTokenGenerator is used for testing that Authenticate calls its provided
// token generator in a specific way.
type mockTokenGenerator struct{}

// GenerateToken implements the TokenGenerator interface. It returns a "token"
// that includes some information about the claims it was passed.
func (mockTokenGenerator) GenerateToken(claims auth.Claims) (string, error) {
	return fmt.Sprintf("sub:%q iss:%d", claims.Subject, claims.IssuedAt), nil
}

// TestAuthenticate validates the behavior around authenticating users.
func TestAuthenticate(t *testing.T) {
	defer tests.Recover(t)

	t.Log("Given the need to authenticate users")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()

			dbConn := test.MasterDB.Copy()
			defer dbConn.Close()

			nu := user.NewUser{
				Name:            "Anna Walker",
				Email:           "anna@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "goroutines",
				PasswordConfirm: "goroutines",
			}

			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			u, err := user.Create(ctx, dbConn, &nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			var tknGen mockTokenGenerator
			tkn, err := user.Authenticate(ctx, dbConn, tknGen, now, "anna@ardanlabs.com", "goroutines")
			if err != nil {
				t.Fatalf("\t%s\tShould be able to generate a token : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to generate a token.", tests.Success)

			want := fmt.Sprintf("sub:%q iss:1538352000", u.ID.Hex())
			if tkn.Token != want {
				t.Log("\t\tGot :", tkn.Token)
				t.Log("\t\tWant:", want)
				t.Fatalf("\t%s\tToken should indicate the specified user and time were used.", tests.Failed)
			}
			t.Logf("\t%s\tToken should indicate the specified user and time were used.", tests.Success)

			if err := user.Delete(ctx, dbConn, u.ID.Hex()); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)
		}
	}
}
