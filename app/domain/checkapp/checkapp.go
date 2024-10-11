// Package checkapp maintains the app layer api for the check domain.
package checkapp

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

type app struct {
	build string
	log   *logger.Logger
	db    *sqlx.DB
}

func newApp(build string, log *logger.Logger, db *sqlx.DB) *app {
	return &app{
		build: build,
		log:   log,
		db:    db,
	}
}

// readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (a *app) readiness(ctx context.Context, r *http.Request) web.Encoder {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := sqldb.StatusCheck(ctx, a.db); err != nil {
		a.log.Info(ctx, "readiness failure", "ERROR", err)
		return errs.New(errs.Internal, err)
	}

	return nil
}

// liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (a *app) liveness(ctx context.Context, r *http.Request) web.Encoder {
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
