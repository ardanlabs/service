// Package checkapi maintains the web based api for system access.
package checkapi

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/checkapp"
)

type api struct {
	checkApp *checkapp.App
}

func newAPI(checkApp *checkapp.App) *api {
	return &api{
		checkApp: checkApp,
	}
}

// readiness checks if the database is ready and if not will return a 500 status.
// Do not respond by just returning an error because further up in the call
// stack it will interpret that as a non-trusted error.
func (api *api) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	if err := api.checkApp.Readiness(ctx); err != nil {
		return nil, errs.Newf(errs.Internal, "database not ready")
	}

	data := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	return data, nil
}

// liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (api *api) liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
	info := api.checkApp.Liveness()

	return info, nil
}
