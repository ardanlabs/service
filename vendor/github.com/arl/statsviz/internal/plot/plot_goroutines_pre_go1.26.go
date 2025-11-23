//go:build !go1.26

package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/sched/goroutines:goroutines",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			goroutines := samples[idx_sched_goroutines_goroutines].Value.Uint64()
			return []uint64{goroutines}
		}
	},

	layout: Scatter{
		Name:  "goroutines",
		Tags:  []tag{tagScheduler},
		Title: "Goroutines",
		Type:  "scatter",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title: "goroutines",
			},
		},
		Subplots: []Subplot{
			{Name: "goroutines", Unitfmt: "%{y}"},
		},
		InfoText: `<i>Goroutines</i> is <b>/sched/goroutines</b>, the count of live goroutines`,
	},
})
