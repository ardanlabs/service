//go:build go1.26

package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/cleanups/executed:cleanups",
		"/gc/cleanups/queued:cleanups",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			executed := samples[idx_gc_cleanups_executed_cleanups].Value.Uint64()
			queued := samples[idx_gc_cleanups_queued_cleanups].Value.Uint64()
			return []uint64{queued - executed}
		}
	},
	layout: Scatter{
		Name:  "gc-cleanups",
		Tags:  []tag{tagGC},
		Title: "GC Cleanups Queue",
		Type:  "bar",
		Layout: ScatterLayout{
			BarMode: "stack",
			Yaxis: ScatterYAxis{
				Title: "cleanups",
			},
		},
		Subplots: []Subplot{
			{Unitfmt: "%{y}", Type: "bar", Name: "queue size"},
		},
		InfoText: `Approximate length of the cleanup functions queue (created by runtime.AddCleanup).
Its <i>/gc/cleanups/queued:cleanups</i> - <i>/gc/cleanups/executed:cleanups</i>.
Useful for detecting slow cleanups holding up the queue.
`},
})
