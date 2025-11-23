package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/heap/allocs:objects",
		"/gc/heap/frees:objects",
	},
	getvalues: func() getvalues {
		rateallocs := rate[uint64]()
		ratefrees := rate[uint64]()

		return func(now time.Time, samples []metrics.Sample) any {
			curallocs := samples[idx_gc_heap_allocs_objects].Value.Uint64()
			curfrees := samples[idx_gc_heap_frees_objects].Value.Uint64()

			return []float64{
				rateallocs(now, curallocs),
				ratefrees(now, curfrees),
			}
		}
	},
	layout: Scatter{
		Name:  "alloc-free-rate",
		Tags:  []tag{tagGC},
		Title: "Heap Allocation & Free Rates",
		Type:  "scatter",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title: "objects / second",
			},
		},
		Subplots: []Subplot{
			{
				Name:    "allocs/sec",
				Unitfmt: "%{y:.4s}",
				Color:   RGBString(66, 133, 244),
			},
			{
				Name:    "frees/sec",
				Unitfmt: "%{y:.4s}",
				Color:   RGBString(219, 68, 55),
			},
		},
		InfoText: `
<i>Allocations per second</i> is the rate of change, per second, of the cumulative <b>/gc/heap/allocs:objects</b> metric.
<i>Frees per second</i> is the rate of change, per second, of the cumulative <b>/gc/heap/frees:objects</b> metric.`,
	},
})
