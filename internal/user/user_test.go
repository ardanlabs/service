package user_test

import (
	"os"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/user"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
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

	t.Log("Given the need to work with Product records.")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()

			dbConn := test.MasterDB.Copy()
			defer dbConn.Close()

			nu := user.NewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			u, err := user.Create(ctx, dbConn, &nu, time.Now().UTC())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			savedU, err := user.Retrieve(ctx, dbConn, u.ID.Hex())
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

			if err := user.Update(ctx, dbConn, u.ID.Hex(), &upd, time.Now().UTC()); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", tests.Success)

			savedU, err = user.Retrieve(ctx, dbConn, u.ID.Hex())
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", tests.Success)

			if savedU.Name != *upd.Name {
				t.Log("\t\tGot :", savedU.Name)
				t.Log("\t\tWant:", *upd.Name)
				t.Errorf("\t%s\tShould be able to see updates to LastName.", tests.Failed)
			} else {
				t.Logf("\t%s\tShould be able to see updates to LastName.", tests.Success)
			}

			if err := user.Delete(ctx, dbConn, u.ID.Hex()); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)

			savedU, err = user.Retrieve(ctx, dbConn, u.ID.Hex())
			if errors.Cause(err) != user.ErrNotFound {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", tests.Success)
		}
	}
}
