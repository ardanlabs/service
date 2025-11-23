package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/stack/starting-size:bytes",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			stackSize := samples[idx_gc_stack_starting_size_bytes].Value.Uint64()
			return []uint64{stackSize}
		}
	},
	layout: Scatter{
		Name:  "gc-stack-size",
		Tags:  []tag{tagGC},
		Title: "Goroutines stack starting size",
		Type:  "scatter",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title: "bytes",
			},
		},
		Subplots: []Subplot{
			{Name: "new goroutines stack size", Unitfmt: "%{y:.4s}B"},
		},
		InfoText: "Shows the stack size of new goroutines, uses <b>/gc/stack/starting-size:bytes</b>",
	},
})
