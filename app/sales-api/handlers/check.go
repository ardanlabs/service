package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/foundation/database"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/api/global"
)

type check struct {
	build string
	db    *sqlx.DB
}

// If the database is not ready we will tell the client and use a 500
// status. Do not respond by just returning an error because further up in
// the call stack will interpret that as a non-trusted error.
func (c *check) health(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.check.health")
	defer span.End()

	status := "ok"
	statusCode := http.StatusOK
	if err := database.StatusCheck(ctx, c.db); err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
	}

	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: c.build,
		Status:  status,
	}
	return web.Respond(ctx, w, health, statusCode)
}
