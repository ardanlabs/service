package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/cgo/go-to-c-calls:calls",
	},
	getvalues: func() getvalues {
		// TODO also show cgo calls per second ?
		deltacalls := delta[uint64]()

		return func(_ time.Time, samples []metrics.Sample) any {
			calls := samples[idx_cgo_go_to_c_calls_calls].Value.Uint64()

			return []uint64{deltacalls(calls)}
		}
	},
	layout: Scatter{
		Name:  "cgo",
		Tags:  []string{tagMisc},
		Title: "CGO Calls",
		Type:  "bar",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title: "calls",
			},
		},
		Subplots: []Subplot{
			{
				Name:    "calls from go to c",
				Unitfmt: "%{y}",
				Color:   "red",
			},
		},
		InfoText: "Shows the count of calls made from Go to C by the current process, per unit of time. Uses <b>/cgo/go-to-c-calls:calls</b>",
	},
})
