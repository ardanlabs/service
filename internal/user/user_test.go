package user_test

import (
	"os"
	"testing"

	"github.com/ardanlabs/service/internal/platform/tests"
	"github.com/ardanlabs/service/internal/user"
)

// TestMain is the entry point for testing.
func TestMain(m *testing.M) {
	os.Exit(tests.Main(m))
}

// TestCreate validates we can create a user and it exists in the DB.
func TestUser(t *testing.T) {
	t.Log("Given the need to validate CRUDing a user.")
	{
		t.Log("\tWhen handling a single user.")
		{
			ctx := tests.Context()

			dbConn, err := tests.MasterDB.Copy()
			if err != nil {
				t.Fatalf("\t%s\tShould be able to connect to mongo : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to connect to mongo.", tests.Success)
			defer dbConn.Close()

			cu := user.CreateUser{
				UserType:  1,
				FirstName: "bill",
				LastName:  "kennedy",
				Email:     "bill@ardanlabs.com",
				Company:   "ardan",
			}

			newUsr, err := user.Create(ctx, dbConn, &cu)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			if _, err = user.Retrieve(ctx, dbConn, newUsr.UserID); err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", tests.Success)

			cu = user.CreateUser{
				UserType:  1,
				FirstName: "bill",
				LastName:  "smith",
				Email:     "bill@ardanlabs.com",
				Company:   "ardan",
			}

			if err := user.Update(ctx, dbConn, newUsr.UserID, &cu); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", tests.Success)

			rtv, err := user.Retrieve(ctx, dbConn, newUsr.UserID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", tests.Success)

			if rtv.LastName != cu.LastName {
				t.Errorf("\t%s\tShould be able to see updates to LastName.", tests.Failed)
				t.Log("\t\tGot :", rtv.LastName)
				t.Log("\t\tWant:", cu.LastName)
			} else {
				t.Logf("\t%s\tShould be able to see updates to LastName.", tests.Success)
			}

			if err := user.Delete(ctx, dbConn, newUsr.UserID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)

			if _, err := user.Retrieve(ctx, dbConn, newUsr.UserID); err == nil {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", tests.Success)

			// TODO: Figure this compare out.
			// if !reflect.DeepEqual(newUsr, rtvUsr) {
			// 	t.Fatalf("\t%s\tShould get back the same user.", Failed)
			// }
			// t.Logf("\t%s\tShould get back the same user.", Success)
		}
	}
}
