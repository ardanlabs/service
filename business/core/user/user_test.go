package user_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/event"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/user/stores/usercache"
	"github.com/ardanlabs/service/business/core/user/stores/userdb"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_User(t *testing.T) {
	t.Run("crud", crud)
	t.Run("paging", paging)
}

func crud(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		teardown()
	}()

	envCore := event.NewCore(log)
	usrCore := user.NewCore(envCore, usercache.NewStore(log, userdb.NewStore(log, db)))

	t.Log("Given the need to work with User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single User.", testID)
		{
			ctx := context.Background()

			email, err := mail.ParseAddress("bill@ardanlabs.com")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse email: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse email.", dbtest.Success, testID)

			nu := user.NewUser{
				Name:            "Bill Kennedy",
				Email:           *email,
				Roles:           []user.Role{user.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			usr, err := usrCore.Create(ctx, nu)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create user.", dbtest.Success, testID)

			saved, err := usrCore.QueryByID(ctx, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by ID.", dbtest.Success, testID)

			if usr.DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.DateCreated)
				t.Logf("\t\tTest %d:\tExp: %v", testID, usr.DateCreated)
				t.Logf("\t\tTest %d:\tDiff: %v", testID, saved.DateCreated.Sub(usr.DateCreated))
				t.Fatalf("\t%s\tTest %d:\tShould get back the same date created.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same date created.", dbtest.Success, testID)

			if usr.DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.DateUpdated)
				t.Logf("\t\tTest %d:\tExp: %v", testID, usr.DateUpdated)
				t.Logf("\t\tTest %d:\tDiff: %v", testID, saved.DateUpdated.Sub(usr.DateUpdated))
				t.Fatalf("\t%s\tTest %d:\tShould get back the same date updated.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same date updated.", dbtest.Success, testID)

			usr.DateCreated = time.Time{}
			usr.DateUpdated = time.Time{}
			saved.DateCreated = time.Time{}
			saved.DateUpdated = time.Time{}

			if diff := cmp.Diff(usr, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same user.", dbtest.Success, testID)

			email, err = mail.ParseAddress("jacob@ardanlabs.com")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse email: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse email.", dbtest.Success, testID)

			upd := user.UpdateUser{
				Name:       dbtest.StringPointer("Jacob Walker"),
				Email:      email,
				Department: dbtest.StringPointer("development"),
			}

			if _, err := usrCore.Update(ctx, saved, upd); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update user.", dbtest.Success, testID)

			saved, err = usrCore.QueryByEmail(ctx, *upd.Email)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by Email : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by Email.", dbtest.Success, testID)

			diff := usr.DateUpdated.Sub(saved.DateUpdated)
			if diff > 0 {
				t.Fatalf("Should have a larger DateUpdated : sav %v, usr %v, dif %v", saved.DateUpdated, usr.DateUpdated, diff)
			}

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

			if saved.Department != *upd.Department {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Department.", dbtest.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Department)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Department)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Department.", dbtest.Success, testID)
			}

			if err := usrCore.Delete(ctx, saved); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete user.", dbtest.Success, testID)

			_, err = usrCore.QueryByID(ctx, saved.ID)
			if !errors.Is(err, user.ErrNotFound) {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve user.", dbtest.Success, testID)
		}
	}
}

func paging(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		teardown()
	}()

	envCore := event.NewCore(log)
	usrCore := user.NewCore(envCore, usercache.NewStore(log, userdb.NewStore(log, db)))

	t.Log("Given the need to page through User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen paging through 2 users.", testID)
		{
			ctx := context.Background()

			name := "User Gopher"
			users1, err := usrCore.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user %q.", dbtest.Success, testID, name)

			n, err := usrCore.Count(ctx, user.QueryFilter{Name: &name})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user count %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user count %q.", dbtest.Success, testID, name)

			if len(users1) != n && users1[0].Name == name {
				t.Fatalf("\t%s\tTest %d:\tShould have a single user for %q", dbtest.Failed, testID, name)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single user.", dbtest.Success, testID)

			name = "Admin Gopher"
			users2, err := usrCore.Query(ctx, user.QueryFilter{Name: &name}, user.DefaultOrderBy, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve users %q.", dbtest.Success, testID, name)

			n, err = usrCore.Count(ctx, user.QueryFilter{Name: &name})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user count %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user count %q.", dbtest.Success, testID, name)

			if len(users2) != n && users2[0].Name == name {
				t.Fatalf("\t%s\tTest %d:\tShould have a single user for %q.", dbtest.Failed, testID, name)
			}
			t.Logf("\t%s\tTest %d:\tShould have a single user.", dbtest.Success, testID)

			users3, err := usrCore.Query(ctx, user.QueryFilter{}, user.DefaultOrderBy, 1, 2)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve 2 users for page 1 : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve 2 users for page 1.", dbtest.Success, testID)

			n, err = usrCore.Count(ctx, user.QueryFilter{})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user count %q : %s.", dbtest.Failed, testID, name, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user count %q.", dbtest.Success, testID, name)

			if len(users3) != n {
				t.Logf("\t\tTest %d:\tgot: %v", testID, len(users3))
				t.Logf("\t\tTest %d:\texp: %v", testID, n)
				t.Fatalf("\t%s\tTest %d:\tShould have 2 users for page 1.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould have 2 users for page 1.", dbtest.Success, testID)

			if users3[0].ID == users3[1].ID {
				t.Logf("\t\tTest %d:\tUser1: %v", testID, users3[0].ID)
				t.Logf("\t\tTest %d:\tUser2: %v", testID, users3[1].ID)
				t.Fatalf("\t%s\tTest %d:\tShould have different users.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould have different users.", dbtest.Success, testID)
		}
	}
}
