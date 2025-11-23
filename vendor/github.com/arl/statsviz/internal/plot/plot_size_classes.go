package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/heap/allocs-by-size:bytes",
		"/gc/heap/frees-by-size:bytes",
	},
	getvalues: func() getvalues {
		var sizeClasses []uint64

		return func(_ time.Time, samples []metrics.Sample) any {
			allocsBySize := samples[idx_gc_heap_allocs_by_size_bytes].Value.Float64Histogram()
			freesBySize := samples[idx_gc_heap_frees_by_size_bytes].Value.Float64Histogram()

			if sizeClasses == nil {
				sizeClasses = make([]uint64, len(allocsBySize.Counts))
			}

			for i := range sizeClasses {
				sizeClasses[i] = allocsBySize.Counts[i] - freesBySize.Counts[i]
			}

			return sizeClasses
		}
	},
	layout: func(samples []metrics.Sample) Heatmap {
		// Perform a sanity check on the number of buckets on the 'allocs' and
		// 'frees' size classes histograms. Statsviz plots a single histogram based
		// on those 2 so we want them to have the same number of buckets, which
		// should be true.
		allocsBySize := samples[idx_gc_heap_allocs_by_size_bytes].Value.Float64Histogram()
		freesBySize := samples[idx_gc_heap_frees_by_size_bytes].Value.Float64Histogram()
		if len(allocsBySize.Buckets) != len(freesBySize.Buckets) {
			panic("different number of buckets in allocs and frees size classes histograms")
		}

		// No downsampling for the size classes histogram (factor=1) but we still
		// need to adapt boundaries for plotly heatmaps.
		buckets := downsampleBuckets(allocsBySize, 1)

		return Heatmap{
			Name:       "size-classes",
			Tags:       []tag{tagGC},
			Title:      "Size Classes",
			Type:       "heatmap",
			UpdateFreq: 5,
			Colorscale: BlueShades,
			Buckets:    floatseq(len(buckets)),
			CustomData: buckets,
			Hover: HeapmapHover{
				YName: "size class",
				YUnit: "bytes",
				ZName: "objects",
			},
			InfoText: `This heatmap shows the distribution of size classes, using <b>/gc/heap/allocs-by-size</b> and <b>/gc/heap/frees-by-size</b>.`,
			Layout: HeatmapLayout{
				YAxis: HeatmapYaxis{
					Title:    "size class",
					TickMode: "array",
					TickVals: []float64{1, 9, 17, 25, 31, 37, 43, 50, 58, 66},
					TickText: []float64{1 << 4, 1 << 7, 1 << 8, 1 << 9, 1 << 10, 1 << 11, 1 << 12, 1 << 13, 1 << 14, 1 << 15},
				},
			},
		}
	},
})
