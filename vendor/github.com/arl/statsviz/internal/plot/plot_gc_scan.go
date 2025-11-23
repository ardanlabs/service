package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/scan/globals:bytes",
		"/gc/scan/heap:bytes",
		"/gc/scan/stack:bytes",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			globals := samples[idx_gc_scan_globals_bytes].Value.Uint64()
			heap := samples[idx_gc_scan_heap_bytes].Value.Uint64()
			stack := samples[idx_gc_scan_stack_bytes].Value.Uint64()

			return []uint64{globals, heap, stack}
		}
	},
	layout: Scatter{
		Name:   "gc-scan",
		Tags:   []tag{tagGC},
		Title:  "GC Scan",
		Type:   "bar",
		Events: "lastgc",
		Layout: ScatterLayout{
			BarMode: "stack",
			Yaxis: ScatterYAxis{
				TickSuffix: "B",
				Title:      "bytes",
			},
		},
		Subplots: []Subplot{
			{
				Name:    "scannable globals",
				Unitfmt: "%{y:.4s}B",
				Type:    "bar",
			},
			{
				Name:    "scannable heap",
				Unitfmt: "%{y:.4s}B",
				Type:    "bar",
			},
			{
				Name:    "scanned stack",
				Unitfmt: "%{y:.4s}B",
				Type:    "bar",
			},
		},
		InfoText: `
This plot shows the amount of memory that is scannable by the GC.
<i>scannable globals</i> is <b>/gc/scan/globals</b>, the total amount of global variable space that is scannable.
<i>scannable heap</i> is <b>/gc/scan/heap</b>, the total amount of heap space that is scannable.
<i>scanned stack</i> is <b>/gc/scan/stack</b>, the number of bytes of stack that were scanned last GC cycle.
`,
	},
})
