package handlers

import (
	"context"
	"net/http"
	"os"

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

// The Info route returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (c *check) info(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := "up"
	statusCode := http.StatusOK

	host, _ := os.Hostname()
	info := struct {
		Status    string `json:"status,omitempty"`
		Host      string `json:"host,omitempty"`
		PodIP     string `json:"pod_ip,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    status,
		Host:      host,
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	return web.Respond(ctx, w, info, statusCode)
}
