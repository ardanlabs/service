package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/web"
)

// Health provides support for orchestration health checks.
type Health struct{}

// Check validates the service is ready and healthy to accept requests.
func (h *Health) Check(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	status := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	web.Respond(ctx, w, status, http.StatusOK)
	return nil
}
