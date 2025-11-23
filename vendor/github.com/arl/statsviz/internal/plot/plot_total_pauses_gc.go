package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/sched/pauses/total/gc:seconds",
	},
	getvalues: func() getvalues {
		histfactor := 0
		counts := [maxBuckets]uint64{}

		return func(_ time.Time, samples []metrics.Sample) any {
			hist := samples[idx_sched_pauses_total_gc_seconds].Value.Float64Histogram()
			if histfactor == 0 {
				histfactor = downsampleFactor(len(hist.Buckets), maxBuckets)
			}

			return downsampleCounts(hist, histfactor, counts[:])
		}
	},
	layout: func(samples []metrics.Sample) Heatmap {
		hist := samples[idx_sched_pauses_total_gc_seconds].Value.Float64Histogram()
		histfactor := downsampleFactor(len(hist.Buckets), maxBuckets)
		buckets := downsampleBuckets(hist, histfactor)

		return Heatmap{
			Name:       "total-pauses-gc",
			Tags:       []tag{tagScheduler, tagGC},
			Title:      "Stop-the-world Pause Latencies (Total)",
			Type:       "heatmap",
			UpdateFreq: 5,
			Colorscale: PinkShades,
			Buckets:    floatseq(len(buckets)),
			CustomData: buckets,
			Hover: HeapmapHover{
				YName: "pause duration",
				YUnit: "duration",
				ZName: "pauses",
			},
			Layout: HeatmapLayout{
				YAxis: HeatmapYaxis{
					Title:    "pause duration",
					TickMode: "array",
					TickVals: []float64{6, 13, 20, 26, 33, 39.5, 46, 53, 60, 66, 73, 79, 86},
					TickText: []float64{1e-7, 1e-6, 1e-5, 1e-4, 1e-3, 5e-3, 1e-2, 5e-2, 1e-1, 5e-1, 1, 5, 10},
				},
			},
			InfoText: `This heatmap shows the distribution of individual <b>GC-related</b> stop-the-world <i>pause latencies</i>.
This is the time from deciding to stop the world until the world is started again.
Some of this time is spent getting all threads to stop (this is measured directly in <i>/sched/pauses/stopping/gc:seconds</i>), during which some threads may still be running.
Uses <b>/sched/pauses/total/gc:seconds</b>.`,
		}
	},
})
