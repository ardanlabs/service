package handlers

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/web"
)

// Check provides support for orchestration health checks.
type Check struct{}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	status := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
