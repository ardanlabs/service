package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/auth"
	"github.com/ardanlabs/service/internal/data"
	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestData(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	t.Run("user", user(db))
	t.Run("product", product(db))
	t.Run("authenticate", authenticate(db))
}

func user(db *sqlx.DB) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to work with User records.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a single User.", testID)
			{
				ctx := tests.Context()
				now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

				nu := data.NewUser{
					Name:            "Bill Kennedy",
					Email:           "bill@ardanlabs.com",
					Roles:           []string{auth.RoleAdmin},
					Password:        "gophers",
					PasswordConfirm: "gophers",
				}

				if err := data.DeleteAll(db); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to delete all data : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to delete all data.", tests.Success, testID)

				u, err := data.Create.User(ctx, db, nu, now)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create user.", tests.Success, testID)

				claims := auth.Claims{
					StandardClaims: jwt.StandardClaims{
						Issuer:    "service project",
						Subject:   "718ffbea-f4a1-4667-8ae3-b349da52675e",
						Audience:  "students",
						ExpiresAt: now.Add(time.Hour).Unix(),
						IssuedAt:  now.Unix(),
					},
					Roles: []string{auth.RoleAdmin, auth.RoleUser},
				}

				savedU, err := data.Retrieve.User.One(ctx, claims, db, u.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by ID: %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by ID.", tests.Success, testID)

				if diff := cmp.Diff(u, savedU); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same user.", tests.Success, testID)

				upd := data.UpdateUser{
					Name:  tests.StringPointer("Jacob Walker"),
					Email: tests.StringPointer("jacob@ardanlabs.com"),
				}

				if err := data.Update.User(ctx, claims, db, u.ID, upd, now); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to update user : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to update user.", tests.Success, testID)

				savedU, err = data.Retrieve.User.One(ctx, claims, db, u.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to retrieve user.", tests.Success, testID)

				if savedU.Name != *upd.Name {
					t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Name.", tests.Failed, testID)
					t.Logf("\t\tTest %d:\tGot: %v", testID, savedU.Name)
					t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Name)
				} else {
					t.Logf("\t%s\tTest %d:\tShould be able to see updates to Name.", tests.Success, testID)
				}

				if savedU.Email != *upd.Email {
					t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Email.", tests.Failed, testID)
					t.Logf("\t\tTest %d:\tGot: %v", testID, savedU.Email)
					t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Email)
				} else {
					t.Logf("\t%s\tTest %d:\tShould be able to see updates to Email.", tests.Success, testID)
				}

				if err := data.Delete.User(ctx, db, u.ID); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to delete user : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to delete user.", tests.Success, testID)

				_, err = data.Retrieve.User.One(ctx, claims, db, u.ID)
				if errors.Cause(err) != data.ErrNotFound {
					t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve user : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve user.", tests.Success, testID)
			}
		}
	}
	return tf
}

func product(db *sqlx.DB) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to work with Product records.")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a single Product.", testID)
			{
				np := data.NewProduct{
					Name:     "Comic Books",
					Cost:     10,
					Quantity: 55,
				}
				now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
				ctx := context.Background()

				if err := data.DeleteAll(db); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to delete all data : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to delete all data.", tests.Success, testID)

				claims := auth.Claims{
					StandardClaims: jwt.StandardClaims{
						Issuer:    "service project",
						Subject:   "718ffbea-f4a1-4667-8ae3-b349da52675e",
						Audience:  "students",
						ExpiresAt: now.Add(time.Hour).Unix(),
						IssuedAt:  now.Unix(),
					},
					Roles: []string{auth.RoleAdmin, auth.RoleUser},
				}

				p, err := data.Create.Product(ctx, db, claims, np, now)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create a product : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create a product.", tests.Success, testID)

				saved, err := data.Retrieve.Product.One(ctx, db, p.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve product by ID: %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to retrieve product by ID.", tests.Success, testID)

				if diff := cmp.Diff(p, saved); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

				upd := data.UpdateProduct{
					Name:     tests.StringPointer("Comics"),
					Cost:     tests.IntPointer(50),
					Quantity: tests.IntPointer(40),
				}
				updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

				if err := data.Update.Product(ctx, db, claims, p.ID, upd, updatedTime); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to update product : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to update product.", tests.Success, testID)

				saved, err = data.Retrieve.Product.One(ctx, db, p.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

				// Check specified fields were updated. Make a copy of the original product
				// and change just the fields we expect then diff it with what was saved.
				want := *p
				want.Name = *upd.Name
				want.Cost = *upd.Cost
				want.Quantity = *upd.Quantity
				want.DateUpdated = updatedTime

				if diff := cmp.Diff(want, *saved); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same product. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same product.", tests.Success, testID)

				upd = data.UpdateProduct{
					Name: tests.StringPointer("Graphic Novels"),
				}

				if err := data.Update.Product(ctx, db, claims, p.ID, upd, updatedTime); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of product : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of product.", tests.Success, testID)

				saved, err = data.Retrieve.Product.One(ctx, db, p.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated product : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated product.", tests.Success, testID)

				if saved.Name != *upd.Name {
					t.Fatalf("\t%s\tTest %d:\tShould be able to see updated Name field : got %q want %q.", tests.Failed, testID, saved.Name, *upd.Name)
				} else {
					t.Logf("\t%s\tTest %d:\tShould be able to see updated Name field.", tests.Success, testID)
				}

				if err := data.Delete.Product(ctx, db, p.ID); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to delete product : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to delete product.", tests.Success, testID)

				_, err = data.Retrieve.Product.One(ctx, db, p.ID)
				if errors.Cause(err) != data.ErrNotFound {
					t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted product.", tests.Success, testID)
			}
		}
	}
	return tf
}

func authenticate(db *sqlx.DB) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to authenticate users")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a single User.", testID)
			{
				ctx := tests.Context()

				nu := data.NewUser{
					Name:            "Anna Walker",
					Email:           "anna@ardanlabs.com",
					Roles:           []string{auth.RoleAdmin},
					Password:        "goroutines",
					PasswordConfirm: "goroutines",
				}

				now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

				if err := data.DeleteAll(db); err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to delete all data : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to delete all data.", tests.Success, testID)

				u, err := data.Create.User(ctx, db, nu, now)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create user.", tests.Success, testID)

				claims, err := data.Authenticate(ctx, db, now, "anna@ardanlabs.com", "goroutines")
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to generate claims : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to generate claims.", tests.Success, testID)

				want := auth.Claims{
					Roles: u.Roles,
					StandardClaims: jwt.StandardClaims{
						Issuer:    "service project",
						Subject:   u.ID,
						Audience:  "students",
						ExpiresAt: now.Add(time.Hour).Unix(),
						IssuedAt:  now.Unix(),
					},
				}

				if diff := cmp.Diff(want, claims); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the expected claims. Diff:\n%s", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the expected claims.", tests.Success, testID)
			}
		}
	}
	return tf
}
