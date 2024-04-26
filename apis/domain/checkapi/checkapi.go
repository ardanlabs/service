// Package checkapi maintains the web based api for system access.
package checkapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/domain/checkapp"
	"github.com/ardanlabs/service/foundation/web"
)

type api struct {
	checkApp *checkapp.Core
}

func newAPI(checkApp *checkapp.Core) *api {
	return &api{
		checkApp: checkApp,
	}
}

// readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (api *api) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := "ok"
	statusCode := http.StatusOK

	if err := api.checkApp.Readiness(ctx); err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
	}

	data := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	return web.Respond(ctx, w, data, statusCode)
}

// liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (api *api) liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	info := api.checkApp.Liveness()

	return web.Respond(ctx, w, info, http.StatusOK)
}
