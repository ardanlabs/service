// Package debug provides handler support for the debugging endpoints.
package debug

import (
	"context"
	"expvar"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/ardanlabs/service/foundation/logger"
	"github.com/arl/statsviz"
)

// Mux registers all the debug routes from the standard library into a new mux
// bypassing the use of the DefaultServerMux. Using the DefaultServerMux would
// be a security risk since a dependency could inject a handler into our service
// without us knowing it.
func Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars/", expvar.Handler())

	statsviz.Register(mux)

	return mux
}

// LogBuildInfo logs information stored inside the Go binary.
func LogBuildInfo(ctx context.Context, log *logger.Logger) {
	var values []any

	info, _ := debug.ReadBuildInfo()

	for _, s := range info.Settings {
		key := s.Key
		if quoteKey(key) {
			key = strconv.Quote(key)
		}

		value := s.Value
		if quoteValue(value) {
			value = strconv.Quote(value)
		}

		values = append(values, key, value)
	}

	values = append(values, "goversion", info.GoVersion)
	values = append(values, "modversion", info.Main.Version)

	log.Info(ctx, "build info", values...)
}

// quoteKey reports whether key is required to be quoted.
func quoteKey(key string) bool {
	return len(key) == 0 || strings.ContainsAny(key, "= \t\r\n\"`")
}

// quoteValue reports whether value is required to be quoted.
func quoteValue(value string) bool {
	return strings.ContainsAny(value, " \t\r\n\"`")
}
