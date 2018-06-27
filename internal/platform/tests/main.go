package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/docker"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pborman/uuid"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

// Test owns state for running/shutting down tests.
type Test struct {
	Log       *log.Logger
	MasterDB  *db.DB
	container *docker.Container
}

// New is the entry point for tests.
func New() *Test {

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "CRUD-TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// ============================================================
	// Startup Mongo container

	container, err := docker.StartMongo(log)
	if err != nil {
		log.Fatalln(err)
	}

	// ============================================================
	// Configuration

	dbDialTimeout := 25 * time.Second
	dbHost := fmt.Sprintf("mongodb://localhost:%s/gotraining", container.Port)

	// ============================================================
	// Start Mongo

	log.Println("main : Started : Initialize Mongo")
	masterDB, err := db.New(dbHost, dbDialTimeout)
	if err != nil {
		log.Fatalf("startup : Register DB : %v", err)
	}

	return &Test{log, masterDB, container}
}

// TearDown is used for shutting down tests. Calling this should be
// done in a defer immediately after calling New.
func (t *Test) TearDown() {
	t.MasterDB.Close()
	if err := docker.StopMongo(t.Log, t.container); err != nil {
		t.Log.Println(err)
	}
}

// Recover is used to prevent panics from allowing the test to cleanup.
func Recover(t *testing.T) {
	if r := recover(); r != nil {
		t.Fatal("Unhandled Exception:", string(debug.Stack()))
	}
}

// Context returns an app level context for testing.
func Context() context.Context {
	values := web.Values{
		TraceID: uuid.New(),
		Now:     time.Now(),
	}

	return context.WithValue(context.Background(), web.KeyValues, &values)
}
