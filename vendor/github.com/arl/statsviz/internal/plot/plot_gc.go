package plot

import (
	"math"
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/gc/gomemlimit:bytes",
		"/gc/heap/live:bytes",
		"/gc/heap/goal:bytes",
		"/memory/classes/total:bytes",
		"/memory/classes/heap/released:bytes",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			memLimit := samples[idx_gc_gomemlimit_bytes].Value.Uint64()
			heapLive := samples[idx_gc_heap_live_bytes].Value.Uint64()
			heapGoal := samples[idx_gc_heap_goal_bytes].Value.Uint64()
			memTotal := samples[idx_memory_classes_total_bytes].Value.Uint64()
			heapReleased := samples[idx_memory_classes_heap_released_bytes].Value.Uint64()

			if memLimit == math.MaxInt64 {
				memLimit = 0
			}

			return []uint64{
				memLimit,
				memTotal - heapReleased,
				heapLive,
				heapGoal,
			}
		}
	},
	layout: Scatter{
		Name:   "garbage collection",
		Tags:   []tag{tagGC},
		Title:  "GC Memory Summary",
		Type:   "scatter",
		Events: "lastgc",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title:      "bytes",
				TickSuffix: "B",
			},
		},
		Subplots: []Subplot{
			{Name: "memory limit", Unitfmt: "%{y:.4s}B"},
			{Name: "in-use memory", Unitfmt: "%{y:.4s}B"},
			{Name: "heap live", Unitfmt: "%{y:.4s}B"},
			{Name: "heap goal", Unitfmt: "%{y:.4s}B"},
		},
		InfoText: `
<i>Memory limit</i> is <b>/gc/gomemlimit:bytes</b>, the Go runtime memory limit configured by the user (via GOMEMLIMIT or debug.SetMemoryLimt), otherwise 0. 
<i>In-use memory</i> is the total mapped memory minus released heap memory (<b>/memory/classes/total - /memory/classes/heap/released</b>).
<i>Heap live</i> is <b>/gc/heap/live:bytes</b>, heap memory occupied by live objects.  
<i>Heap goal</i> is <b>/gc/heap/goal:bytes</b>, the heap size target at the end of each GC cycle.`,
	},
})
