package mid

import (
	"context"
	"net/http"

	"github.com/ardanlabs/service/business/web/metrics"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics updates program counters.
func Metrics() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx = metrics.Set(ctx)

			var err error
			promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = handler(ctx, w, r)
			})).ServeHTTP(w, r)

			metrics.AddRequests(ctx)
			metrics.AddGoroutines(ctx)

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}

		return h
	}

	return m
}
