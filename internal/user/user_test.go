package user_test

import (
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/auth"
	"github.com/ardanlabs/service/internal/platform/database/databasetest"
	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/user"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

// TestUser validates the full set of CRUD operations on User values.
func TestUser(t *testing.T) {
	db, teardown := databasetest.Setup(t)
	defer teardown()

	t.Log("Given the need to work with User records.")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			// claims is information about the person making the request.
			claims := auth.NewClaims(
				"718ffbea-f4a1-4667-8ae3-b349da52675e", // This is just some random UUID.
				[]string{auth.RoleAdmin, auth.RoleUser},
				now, time.Hour,
			)

			nu := user.NewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			u, err := user.Create(ctx, db, &nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			savedU, err := user.Retrieve(ctx, claims, db, u.ID)
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

			if err := user.Update(ctx, claims, db, u.ID, &upd, now); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", tests.Success)

			savedU, err = user.Retrieve(ctx, claims, db, u.ID)
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

			if err := user.Delete(ctx, db, u.ID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)

			savedU, err = user.Retrieve(ctx, claims, db, u.ID)
			if errors.Cause(err) != user.ErrNotFound {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", tests.Success)
		}
	}
}

// TestAuthenticate validates the behavior around authenticating users.
func TestAuthenticate(t *testing.T) {
	db, teardown := databasetest.Setup(t)
	defer teardown()

	t.Log("Given the need to authenticate users")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()

			nu := user.NewUser{
				Name:            "Anna Walker",
				Email:           "anna@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "goroutines",
				PasswordConfirm: "goroutines",
			}

			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			u, err := user.Create(ctx, db, &nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			claims, err := user.Authenticate(ctx, db, now, "anna@ardanlabs.com", "goroutines")
			if err != nil {
				t.Fatalf("\t%s\tShould be able to generate claims : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to generate claims.", tests.Success)

			want := auth.Claims{}
			want.Subject = u.ID
			want.Roles = u.Roles
			want.ExpiresAt = now.Add(time.Hour).Unix()
			want.IssuedAt = now.Unix()

			if diff := cmp.Diff(want, claims); diff != "" {
				t.Fatalf("\t%s\tShould get back the expected claims. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the expected claims.", tests.Success)
		}
	}
}
