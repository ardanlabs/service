//go:build go1.26

package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/finalizers/executed:finalizers",
		"/gc/finalizers/queued:finalizers",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			executed := samples[idx_gc_finalizers_executed_finalizers].Value.Uint64()
			queued := samples[idx_gc_finalizers_queued_finalizers].Value.Uint64()
			return []uint64{queued - executed}
		}
	},
	layout: Scatter{
		Name:  "gc-finalizers",
		Tags:  []tag{tagGC},
		Title: "GC Finalizers Queue",
		Type:  "bar",
		Layout: ScatterLayout{
			BarMode: "stack",
			Yaxis: ScatterYAxis{
				Title: "finalizers",
			},
		},
		Subplots: []Subplot{
			{Unitfmt: "%{y}", Type: "bar", Name: "queue size"},
		},
		InfoText: `Length of the finalizer functions queue (created by runtime.AddFinalizer).
Its <i>/gc/finalizers/queued:cleanups</i> - <i>/gc/finalizers/executed:cleanups</i>.
Useful for detecting finalizers overwhelming the queue, either by being too slow, or by there being too many of them.
`},
})
