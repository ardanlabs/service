package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/database"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"
)

// Check provides support for orchestration health checks.
type Check struct {
	db *sqlx.DB

	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Check.Health")
	defer span.End()

	var status struct {
		Status string `json:"status"`
	}

	// Check if the database is ready.
	if err := database.StatusCheck(ctx, db); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		status.Status = "db not ready"
		return web.Respond(ctx, w, status, http.StatusInternalServerError)
	}

	status.Status = "ok"
	return web.Respond(ctx, w, status, http.StatusOK)
}
