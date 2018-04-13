package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/web"
	"go.opencensus.io/trace"
)

// Health provides support for orchestration health checks.
type Health struct {
	MasterDB *db.DB
}

// Check validates the service is ready and healthy to accept requests.
func (h *Health) Check(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Health.Check")
	defer span.End()

	dbConn := h.MasterDB.Copy()
	defer dbConn.Close()

	if err := dbConn.StatusCheck(ctx); err != nil {
		return err
	}

	status := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	web.Respond(ctx, w, status, http.StatusOK)
	return nil
}
