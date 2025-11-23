package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/sched/latencies:seconds",
	},
	getvalues: func() getvalues {
		histfactor := 0
		counts := [maxBuckets]uint64{}

		return func(_ time.Time, samples []metrics.Sample) any {
			hist := samples[idx_sched_latencies_seconds].Value.Float64Histogram()
			if histfactor == 0 {
				histfactor = downsampleFactor(len(hist.Buckets), maxBuckets)
			}

			return downsampleCounts(hist, histfactor, counts[:])
		}
	},
	layout: func(samples []metrics.Sample) Heatmap {
		hist := samples[idx_sched_latencies_seconds].Value.Float64Histogram()
		histfactor := downsampleFactor(len(hist.Buckets), maxBuckets)
		buckets := downsampleBuckets(hist, histfactor)

		return Heatmap{
			Name:       "runnable-time",
			Tags:       []tag{tagScheduler},
			Title:      "Time Goroutines Spend in 'Runnable' state",
			Type:       "heatmap",
			UpdateFreq: 5,
			Colorscale: GreenShades,
			Buckets:    floatseq(len(buckets)),
			CustomData: buckets,
			Hover: HeapmapHover{
				YName: "duration",
				YUnit: "duration",
				ZName: "goroutines",
			},
			Layout: HeatmapLayout{
				YAxis: HeatmapYaxis{
					Title:    "duration",
					TickMode: "array",
					TickVals: []float64{6, 13, 20, 26, 33, 39.5, 46, 53, 60, 66, 73, 79, 86},
					TickText: []float64{1e-7, 1e-6, 1e-5, 1e-4, 1e-3, 5e-3, 1e-2, 5e-2, 1e-1, 5e-1, 1, 5, 10},
				},
			},
			InfoText: `This heatmap shows the distribution of the time goroutines have spent in the scheduler in a runnable state before actually running, uses <b>/sched/latencies:seconds</b>.`,
		}
	},
})
