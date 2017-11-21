package tests

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ardanlabs/service/cmd/crud/handlers"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/docker"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pborman/uuid"
)

// Success and failure markers.
var (
	Success = "\u2713"
	Failed  = "\u2717"
)

var a *web.App
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

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	a = handlers.API(masterDB).(*web.App)

	return m.Run()
}
