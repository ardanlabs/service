package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/ardanlabs/service/internal/mid"
	"github.com/ardanlabs/service/internal/platform/web"
)

// API returns a handler for a set of routes.
func API(shutdown chan os.Signal, log *log.Logger) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, log, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics())

	// Because our static directory is set as the root of the FileSystem,
	// we need to strip off the /static/ prefix from the request path
	// before searching the FileSystem for the given file.
	fs := http.FileServer(http.Dir("static"))
	sp := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	}
	app.TreeMux.Handle("GET", "/static/*", sp)

	// Register health check endpoint. This route is not authenticated.
	check := Check{}
	app.Handle("GET", "/health", check.Health)

	// Register health check endpoint. This route is not authenticated.
	search := NewSearch(log)
	app.Handle("GET", "/search", search.Query)
	app.Handle("POST", "/search", search.Query)

	return app
}
