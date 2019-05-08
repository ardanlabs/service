package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"go.opencensus.io/trace"
)

// Check provides support for orchestration health checks.
type Check struct {
	MasterDB *db.DB

	// ADD OTHER STATE LIKE THE LOGGER IF NEEDED.
}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Check.Health")
	defer span.End()

	dbConn := c.MasterDB.Copy()
	defer dbConn.Close()

	if err := dbConn.StatusCheck(ctx); err != nil {
		return err
	}

	status := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
