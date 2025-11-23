package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/sched/pauses/stopping/other:seconds",
	},
	getvalues: func() getvalues {
		histfactor := 0
		counts := [maxBuckets]uint64{}

		return func(_ time.Time, samples []metrics.Sample) any {
			hist := samples[idx_sched_pauses_stopping_other_seconds].Value.Float64Histogram()
			if histfactor == 0 {
				histfactor = downsampleFactor(len(hist.Buckets), maxBuckets)
			}

			return downsampleCounts(hist, histfactor, counts[:])
		}
	},
	layout: func(samples []metrics.Sample) Heatmap {
		hist := samples[idx_sched_pauses_stopping_other_seconds].Value.Float64Histogram()
		histfactor := downsampleFactor(len(hist.Buckets), maxBuckets)
		buckets := downsampleBuckets(hist, histfactor)

		return Heatmap{
			Name:       "stopping-pauses-other",
			Tags:       []tag{tagScheduler},
			Title:      "Stop-the-world Stopping Latencies (Other)",
			Type:       "heatmap",
			UpdateFreq: 5,
			Colorscale: GreenShades,
			Buckets:    floatseq(len(buckets)),
			CustomData: buckets,
			Hover: HeapmapHover{
				YName: "stopping duration",
				YUnit: "duration",
				ZName: "pauses",
			},
			Layout: HeatmapLayout{
				YAxis: HeatmapYaxis{
					Title:    "stopping duration",
					TickMode: "array",
					TickVals: []float64{6, 13, 20, 26, 33, 39.5, 46, 53, 60, 66, 73, 79, 86},
					TickText: []float64{1e-7, 1e-6, 1e-5, 1e-4, 1e-3, 5e-3, 1e-2, 5e-2, 1e-1, 5e-1, 1, 5, 10},
				},
			},
			InfoText: `This heatmap shows the distribution of individual <b>non-GC-related</b> stop-the-world <i>stopping latencies</i>.
This is the time it takes from deciding to stop the world until all Ps are stopped.
This is a subset of the total non-GC-related stop-the-world time. During this time, some threads may be executing.
Uses <b>/sched/pauses/stopping/other:seconds</b>.`,
		}
	},
})
