// Package checkapp maintains the app layer api for the check domain.
package checkapp

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/jmoiron/sqlx"
)

// App manages the set of app layer api functions for the check domain.
type App struct {
	build string
	log   *logger.Logger
	db    *sqlx.DB
}

// NewApp constructs a check app API for use.
func NewApp(build string, log *logger.Logger, db *sqlx.DB) *App {
	return &App{
		build: build,
		log:   log,
		db:    db,
	}
}

// Readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (a *App) Readiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := sqldb.StatusCheck(ctx, a.db); err != nil {
		a.log.Info(ctx, "readiness failure", "ERROR", err)
		return err
	}

	return nil
}

// Liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (a *App) Liveness() Info {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	info := Info{
		Status:     "up",
		Build:      a.build,
		Host:       host,
		Name:       os.Getenv("KUBERNETES_NAME"),
		PodIP:      os.Getenv("KUBERNETES_POD_IP"),
		Node:       os.Getenv("KUBERNETES_NODE_NAME"),
		Namespace:  os.Getenv("KUBERNETES_NAMESPACE"),
		GOMAXPROCS: runtime.GOMAXPROCS(0),
	}

	// This handler provides a free timer loop.

	return info
}
