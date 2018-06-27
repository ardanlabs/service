package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(log *log.Logger, zipkinHost string, apiHost string) http.Handler {
	app := web.New(log, mid.RequestLogger, mid.ErrorHandler)

	z := NewZipkin(zipkinHost, apiHost, time.Second)
	app.Handle("POST", "/v1/publish", z.Publish)

	h := Health{}
	app.Handle("GET", "/v1/health", h.Check)

	return app
}
