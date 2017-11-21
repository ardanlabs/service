package user_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ardanlabs/kit/web"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/docker"
	"github.com/ardanlabs/service/internal/user"
	"github.com/pborman/uuid"
)

// TestCreate validates we can create a user and it exists
// in the DB.
func TestUser(t *testing.T) {
	t.Log("Given the need to validate CRUDing a user.")
	{
		t.Log("\tWhen handling a single user.")
		{
			dbConn, err := masterDB.Copy()
			if err != nil {
				t.Fatalf("\t%s\tShould be able to connect to mongo : %s.", Failed, err)
			}
			t.Logf("\t%s\tShould be able to connect to mongo.", Success)
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
				t.Fatalf("\t%s\tShould be able to create user : %s.", Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", Success)

			if _, err = user.Retrieve(ctx, dbConn, newUsr.UserID); err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", Success)

			cu = user.CreateUser{
				UserType:  1,
				FirstName: "bill",
				LastName:  "smith",
				Email:     "bill@ardanlabs.com",
				Company:   "ardan",
			}

			if err := user.Update(ctx, dbConn, newUsr.UserID, &cu); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", Success)

			rtv, err := user.Retrieve(ctx, dbConn, newUsr.UserID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", Success)

			if rtv.LastName != cu.LastName {
				t.Errorf("\t%s\tShould be able to see updates to LastName.", Failed)
				t.Log("\t\tGot :", rtv.LastName)
				t.Log("\t\tWant:", cu.LastName)
			} else {
				t.Logf("\t%s\tShould be able to see updates to LastName.", Success)
			}

			if err := user.Delete(ctx, dbConn, newUsr.UserID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", Success)

			if _, err := user.Retrieve(ctx, dbConn, newUsr.UserID); err == nil {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", Success)

			// TODO: Figure this compare out.
			// if !reflect.DeepEqual(newUsr, rtvUsr) {
			// 	t.Fatalf("\t%s\tShould get back the same user.", Failed)
			// }
			// t.Logf("\t%s\tShould get back the same user.", Success)
		}
	}
}

// =============================================================================

// Success and failure markers.
var (
	Success = "\u2713"
	Failed  = "\u2717"
)

var ctx context.Context
var masterDB *db.DB

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	values := web.Values{
		TraceID: uuid.New(),
		Now:     time.Now(),
	}
	ctx = context.WithValue(context.Background(), web.KeyValues, &values)

	c, err := docker.StartMongo()
	if err != nil {
		log.Fatalln(err)
	}
	docker.SetTestEnv(c)

	defer func() {
		if err := docker.StopMongo(c); err != nil {
			log.Println(err)
		}
	}()

	// TODO: Think about this more.
	dbTimeout := 25 * time.Second
	dbHost := os.Getenv("DB_HOST")

	log.Println("main : Started : Initialize Mongo")
	masterDB, err = db.New(dbHost, dbTimeout)
	if err != nil {
		log.Fatalf("startup : Register DB : %v", err)
	}
	defer masterDB.Close()

	return m.Run()
}
