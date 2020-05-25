package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/database"
	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/api/global"
)

type check struct {
	build string
	db    *sqlx.DB
}

func (c *check) health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.check.health")
	defer span.End()

	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: c.build,
	}

	// Check if the database is ready.
	if err := database.StatusCheck(ctx, c.db); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}

	health.Status = "ok"
	return web.Respond(ctx, w, health, http.StatusOK)
}
