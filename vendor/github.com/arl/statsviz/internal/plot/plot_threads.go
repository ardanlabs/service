//go:build go1.26

package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/sched/threads/total:threads",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			threads := samples[idx_sched_threads_total_threads].Value.Uint64()

			return []uint64{threads}
		}
	},
	layout: Scatter{
		Name:  "threads",
		Tags:  []tag{tagScheduler},
		Title: "Threads",
		Type:  "scatter",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title: "bytes",
			},
		},
		Subplots: []Subplot{
			{
				Name:    "threads",
				Unitfmt: "%{y}",
			},
		},
		InfoText: "Shows the current count of live threads that are owned by the Go runtime. Uses <b>/sched/threads/total:threads</b>",
	},
})
