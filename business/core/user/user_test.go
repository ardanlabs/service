package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/dbschema"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/business/data/dbuser"
	"github.com/ardanlabs/service/business/sys/auth"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-cmp/cmp"
)

var dbc = dbtest.DBContainer{
	Image: "postgres:13-alpine",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
}

func TestUser(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, dbc)
	t.Cleanup(teardown)

	user := user.NewCore(log, db)

	t.Log("Given the need to work with User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single User.", testID)
		{
			ctx := context.Background()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			nu := dbuser.NewUser{
				Name:            "Bill Kennedy",
				Email:           "bill@ardanlabs.com",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			usr, err := user.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create user.", dbtest.Success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   usr.ID,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				Roles: []string{auth.RoleUser},
			}

			saved, err := user.QueryByID(ctx, claims, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by ID.", dbtest.Success, testID)

			if diff := cmp.Diff(usr, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same user.", dbtest.Success, testID)

			upd := dbuser.UpdateUser{
				Name:  dbtest.StringPointer("Jacob Walker"),
				Email: dbtest.StringPointer("jacob@ardanlabs.com"),
			}

			claims = auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				Roles: []string{auth.RoleAdmin},
			}

			if err := user.Update(ctx, claims, usr.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update user.", dbtest.Success, testID)

			saved, err = user.QueryByEmail(ctx, claims, *upd.Email)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by Email : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by Email.", dbtest.Success, testID)

			if saved.Name != *upd.Name {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Name.", dbtest.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Name)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Name.", dbtest.Success, testID)
			}

			if saved.Email != *upd.Email {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Email.", dbtest.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Email)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Email)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Email.", dbtest.Success, testID)
			}

			if err := user.Delete(ctx, claims, usr.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete user.", dbtest.Success, testID)

			_, err = user.QueryByID(ctx, claims, usr.ID)
			if !errors.Is(err, validate.ErrNotFound) {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve user.", dbtest.Success, testID)
		}
	}
}

func TestPagingUser(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, dbc)
	t.Cleanup(teardown)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbschema.Seed(ctx, db)

	user := user.NewCore(log, db)

	t.Log("Given the need to page through User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 users.", testID)
		{
			ctx := context.Background()

			users1, err := user.Query(ctx, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve users for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve users for page 1.", dbtest.Success, testID)

			if len(users1) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single user.", dbtest.Success, testID)

			users2, err := user.Query(ctx, 2, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve users for page 2 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve users for page 2.", dbtest.Success, testID)

			if len(users2) != 1 {
				t.Fatalf("\t%s\tTest %d:\tShould have a single user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single user.", dbtest.Success, testID)

			if users1[0].ID == users2[0].ID {
				t.Logf("\t\tTest %d:\tUser1: %v", testID, users1[0].ID)
				t.Logf("\t\tTest %d:\tUser2: %v", testID, users2[0].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different users : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have different users.", dbtest.Success, testID)
		}
	}
}
