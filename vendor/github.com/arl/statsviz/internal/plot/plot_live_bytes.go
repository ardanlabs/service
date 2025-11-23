package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/heap/allocs:bytes",
		"/gc/heap/frees:bytes",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			allocBytes := samples[idx_gc_heap_allocs_bytes].Value.Uint64()
			freedBytes := samples[idx_gc_heap_frees_bytes].Value.Uint64()

			return []uint64{allocBytes - freedBytes}
		}
	},
	layout: Scatter{
		Name:   "live-bytes",
		Tags:   []tag{tagGC},
		Title:  "Live Bytes in Heap",
		Type:   "bar",
		Events: "lastgc",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title: "bytes",
			},
		},
		Subplots: []Subplot{
			{
				Name:    "live bytes",
				Unitfmt: "%{y:.4s}B",
				Color:   RGBString(135, 182, 218),
			},
		},
		InfoText: `<i>Live bytes</i> is <b>/gc/heap/allocs - /gc/heap/frees</b>. It's the number of bytes currently allocated (and not yet GC'ec) to the heap by the application.`,
	},
})
