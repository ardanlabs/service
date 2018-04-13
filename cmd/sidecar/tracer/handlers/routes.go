package handlers

import (
	"net/http"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API() http.Handler {
	app := web.New(mid.RequestLogger, mid.ErrorHandler)

	s := Span{}
	app.Handle("POST", "/v1/publish", s.Publish)

	h := Health{}
	app.Handle("GET", "/v1/health", h.Check)

	return app
}
