package tests

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ardanlabs/kit/web"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/docker"
	"github.com/pborman/uuid"
)

// Success and failure markers.
var (
	Success = "\u2713"
	Failed  = "\u2717"
)

// MasterDB represents the master Mongo session.
// This will work since we are starting mongo from a container.
var MasterDB *db.DB

// Context returns an app level context for testing.
func Context() context.Context {
	values := web.Values{
		TraceID: uuid.New(),
		Now:     time.Now(),
	}

	return context.WithValue(context.Background(), web.KeyValues, &values)
}

// Main is the entry point for tests.
func Main(m *testing.M) int {

	// ============================================================
	// Startup Mongo container

	c, err := docker.StartMongo()
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err := docker.StopMongo(c); err != nil {
			log.Println(err)
		}
	}()

	// ============================================================
	// Configuration

	dbDialTimeout := 25 * time.Second
	dbHost := fmt.Sprintf("mongodb://localhost:%s/gotraining", c.Port)

	// ============================================================
	// Start Mongo

	log.Println("main : Started : Initialize Mongo")
	MasterDB, err = db.New(dbHost, dbDialTimeout)
	if err != nil {
		log.Fatalf("startup : Register DB : %v", err)
	}
	defer MasterDB.Close()

	// ============================================================
	// Run tests

	return m.Run()
}
