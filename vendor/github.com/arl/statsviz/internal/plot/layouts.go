package plot

import (
	"runtime/metrics"
)

var garbageCollectionLayout = Scatter{
	Name:   "garbage collection",
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
}

var heapDetailslLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "Heap (details)",
	Type:   "scatter",
	Events: "lastgc",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title:      "bytes",
			TickSuffix: "B",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "heap sys",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "heap objects",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "heap stacks",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "heap goal",
			Unitfmt: "%{y:.4s}B",
		},
	},
	InfoText: `
<i>Heap</i> sys is <b>/memory/classes/heap/{objects + unused + released + free}</b>. It's an estimate of all the heap memory obtained from the OS.
<i>Heap objects</i> is <b>/memory/classes/heap/objects</b>, the memory occupied by live objects and dead objects that have not yet been marked free by the GC.
<i>Heap stacks</i> is <b>/memory/classes/heap/stacks</b>, the memory used for stack space.
<i>Heap goal</i> is <b>gc/heap/goal</b>, the heap size target for the end of the GC cycle.`,
}

var liveObjectsLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "Live Objects in Heap",
	Type:   "bar",
	Events: "lastgc",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title: "objects",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "live objects",
			Unitfmt: "%{y:.4s}",
			Color:   RGBString(255, 195, 128),
		},
	},
	InfoText: `<i>Live objects</i> is <b>/gc/heap/objects</b>. It's the number of objects, live or unswept, occupying heap memory.`,
}

var liveBytesLayout = Scatter{
	Name:   "TODO(set later)",
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
}

var mspanMCacheLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "MSpan/MCache",
	Type:   "scatter",
	Events: "lastgc",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title:      "bytes",
			TickSuffix: "B",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "mspan in-use",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "mspan free",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "mcache in-use",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "mcache free",
			Unitfmt: "%{y:.4s}B",
		},
	},
	InfoText: `
<i>Mspan in-use</i> is <b>/memory/classes/metadata/mspan/inuse</b>, the memory that is occupied by runtime mspan structures that are currently being used.
<i>Mspan free</i> is <b>/memory/classes/metadata/mspan/free</b>, the memory that is reserved for runtime mspan structures, but not in-use.
<i>Mcache in-use</i> is <b>/memory/classes/metadata/mcache/inuse</b>, the memory that is occupied by runtime mcache structures that are currently being used.
<i>Mcache free</i> is <b>/memory/classes/metadata/mcache/free</b>, the memory that is reserved for runtime mcache structures, but not in-use.
`,
}

var goroutinesLayout = Scatter{
	Name:  "TODO(set later)",
	Title: "Goroutines",
	Type:  "scatter",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title: "goroutines",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "goroutines",
			Unitfmt: "%{y}",
		},
	},
	InfoText: "<i>Goroutines</i> is <b>/sched/goroutines</b>, the count of live goroutines.",
}

func sizeClassesLayout(samples []metrics.Sample) Heatmap {
	idxallocs := metricIdx["/gc/heap/allocs-by-size:bytes"]
	idxfrees := metricIdx["/gc/heap/frees-by-size:bytes"]

	// Perform a sanity check on the number of buckets on the 'allocs' and
	// 'frees' size classes histograms. Statsviz plots a single histogram based
	// on those 2 so we want them to have the same number of buckets, which
	// should be true.
	allocsBySize := samples[idxallocs].Value.Float64Histogram()
	freesBySize := samples[idxfrees].Value.Float64Histogram()
	if len(allocsBySize.Buckets) != len(freesBySize.Buckets) {
		panic("different number of buckets in allocs and frees size classes histograms")
	}

	// No downsampling for the size classes histogram (factor=1) but we still
	// need to adapt boundaries for plotly heatmaps.
	buckets := downsampleBuckets(allocsBySize, 1)

	return Heatmap{
		Name:       "TODO(set later)",
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
}

func gcPausesLayout(samples []metrics.Sample) Heatmap {
	idxgcpauses := metricIdx["/sched/pauses/total/gc:seconds"]

	gcpauses := samples[idxgcpauses].Value.Float64Histogram()
	histfactor := downsampleFactor(len(gcpauses.Buckets), maxBuckets)
	buckets := downsampleBuckets(gcpauses, histfactor)

	return Heatmap{
		Name:       "TODO(set later)",
		Title:      "Stop-the-world Pause Latencies",
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
		InfoText: `This heatmap shows the distribution of individual GC-related stop-the-world pause latencies, uses <b>/sched/pauses/total/gc:seconds</b>,.`,
	}
}

func runnableTimeLayout(samples []metrics.Sample) Heatmap {
	idxschedlat := metricIdx["/sched/latencies:seconds"]

	schedlat := samples[idxschedlat].Value.Float64Histogram()
	histfactor := downsampleFactor(len(schedlat.Buckets), maxBuckets)
	buckets := downsampleBuckets(schedlat, histfactor)

	return Heatmap{
		Name:       "TODO(set later)",
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
}

var schedEventsLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "Goroutine Scheduling Events",
	Type:   "scatter",
	Events: "lastgc",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title: "events",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "events per unit of time",
			Unitfmt: "%{y}",
		},
		{
			Name:    "events per unit of time, per P",
			Unitfmt: "%{y}",
		},
	},
	InfoText: `<i>Events per second</i> is the sum of all buckets in <b>/sched/latencies:seconds</b>, that is, it tracks the total number of goroutine scheduling events. That number is multiplied by the constant 8.
<i>Events per second per P (processor)</i> is <i>Events per second</i> divided by current <b>GOMAXPROCS</b>, from <b>/sched/gomaxprocs:threads</b>.
<b>NOTE</b>: the multiplying factor comes from internal Go runtime source code and might change from version to version.`,
}

var cgoLayout = Scatter{
	Name:  "TODO(set later)",
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
}

var gcStackSizeLayout = Scatter{
	Name:  "TODO(set later)",
	Title: "Goroutine stack starting size",
	Type:  "scatter",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title: "bytes",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "new goroutines stack size",
			Unitfmt: "%{y:.4s}B",
		},
	},
	InfoText: "Shows the stack size of new goroutines, uses <b>/gc/stack/starting-size:bytes</b>",
}

var gcCyclesLayout = Scatter{
	Name:  "TODO(set later)",
	Title: "Completed GC Cycles",
	Type:  "bar",
	Layout: ScatterLayout{
		BarMode: "stack",
		Yaxis: ScatterYAxis{
			Title: "cycles",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "automatic",
			Unitfmt: "%{y}",
			Type:    "bar",
		},
		{
			Name:    "forced",
			Unitfmt: "%{y}",
			Type:    "bar",
		},
	},
	InfoText: `Number of completed GC cycles, either forced of generated by the Go runtime.`,
}

var memoryClassesLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "Memory classes",
	Type:   "scatter",
	Events: "lastgc",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title:      "bytes",
			TickSuffix: "B",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "os stacks",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "other",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "profiling buckets",
			Unitfmt: "%{y:.4s}B",
		},
		{
			Name:    "total",
			Unitfmt: "%{y:.4s}B",
		},
	},

	InfoText: `
<i>OS stacks</i> is <b>/memory/classes/os-stacks</b>, stack memory allocated by the underlying operating system.
<i>Other</i> is <b>/memory/classes/other</b>, memory used by execution trace buffers, structures for debugging the runtime, finalizer and profiler specials, and more.
<i>Profiling buckets</i> is <b>/memory/classes/profiling/buckets</b>, memory that is used by the stack trace hash map used for profiling.
<i>Total</i> is <b>/memory/classes/total</b>, all memory mapped by the Go runtime into the current process as read-write.`,
}

var cpuGCLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "CPU (Garbage Collector)",
	Type:   "scatter",
	Events: "lastgc",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title:      "cpu-seconds per seconds",
			TickSuffix: "s",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "mark assist",
			Unitfmt: "%{y:.4s}s",
		},
		{
			Name:    "mark dedicated",
			Unitfmt: "%{y:.4s}s",
		},
		{
			Name:    "mark idle",
			Unitfmt: "%{y:.4s}s",
		},
		{
			Name:    "pause",
			Unitfmt: "%{y:.4s}s",
		},
	},

	InfoText: `Cumulative metrics are converted to rates by Statsviz so as to be more easily comparable and readable.
All this metrics are overestimates, and not directly comparable to system CPU time measurements. Compare only with other /cpu/classes metrics.

<i>mark assist</i> is <b>/cpu/classes/gc/mark/assist</b>, estimated total CPU time goroutines spent performing GC tasks to assist the GC and prevent it from falling behind the application.
<i>mark dedicated</i> is <b>/cpu/classes/gc/mark/dedicated</b>, Estimated total CPU time spent performing GC tasks on processors (as defined by GOMAXPROCS) dedicated to those tasks.
<i>mark idle</i> is <b>/cpu/classes/gc/mark/idle</b>, estimated total CPU time spent performing GC tasks on spare CPU resources that the Go scheduler could not otherwise find a use for.
<i>pause</i> is <b>/cpu/classes/gc/pause</b>, estimated total CPU time spent with the application paused by the GC.

All metrics are rates in CPU-seconds per second.`,
}

var cpuScavengerLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "CPU (Scavenger)",
	Type:   "bar",
	Events: "lastgc",
	Layout: ScatterLayout{
		BarMode: "stack",
		Yaxis: ScatterYAxis{
			Title:      "cpu-seconds / second",
			TickSuffix: "s",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "assist",
			Unitfmt: "%{y:.4s}s",
			Type:    "bar",
		},
		{
			Name:    "background",
			Unitfmt: "%{y:.4s}s",
			Type:    "bar",
		},
	},
	InfoText: `Breakdown of how the GC scavenger returns memory to the OS (eagerly vs background).
<i>assist is</i> the rate of <b>/cpu/classes/scavenge/assist</b>, the CPU time spent returning unused memory eagerly in response to memory pressure.
<i>background is</i> the rate of <b>/cpu/classes/scavenge/background</b>, the CPU time spent performing background tasks to return unused memory to the OS.

Both metrics are rates in CPU-seconds per second.`,
}

var cpuOverallLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "CPU (Overall)",
	Type:   "bar",
	Events: "lastgc",
	Layout: ScatterLayout{
		BarMode: "stack",
		Yaxis: ScatterYAxis{
			Title:      "cpu-seconds / second",
			TickSuffix: "s",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "user",
			Unitfmt: "%{y:.4s}s",
			Type:    "bar",
		},
		{
			Name:    "scavenge",
			Unitfmt: "%{y:.4s}s",
			Type:    "bar",
		},
		{
			Name:    "idle",
			Unitfmt: "%{y:.4s}s",
			Type:    "bar",
		},
		{
			Name:    "gc total",
			Unitfmt: "%{y:.4s}s",
			Type:    "bar",
		},
		{
			Name:    "total",
			Unitfmt: "%{y:.4s}s",
			Type:    "scatter",
		},
	},
	InfoText: `Shows the fraction of CPU spent in your code vs. runtime vs. wasted. Helps track overall utilization and potential headroom.
<i>user is</i> the rate of <b>/cpu/classes/user:cpu-seconds</b>, the CPU time spent running user Go code.
<i>scavenge is</i> the rate of <b>/cpu/classes/scavenge:cpu-seconds</b>, the CPU time spent performing tasks that return unused memory to the OS.
<i>idle is</i> the rate of <b>/cpu/classes/idle:cpu-seconds</b>, the CPU time spent performing GC tasks on spare CPU resources that the Go scheduler could not otherwise find a use for.
<i>gc total is</i> the rate of <b>/cpu/classes/gc/total:cpu-seconds</b>, the CPU time spent performing GC tasks (sum of all metrics in <b>/cpu/classes/gc</b>)
<i>total is</i> the rate of <b>/cpu/classes/total:cpu-seconds</b>, the available CPU time for user Go code or the Go runtime, as defined by GOMAXPROCS. In other words, GOMAXPROCS integrated over the wall-clock duration this process has been executing for.

All metrics are rates in CPU-seconds per second.`,
}

var mutexWaitLayout = Scatter{
	Name:   "TODO(set later)",
	Title:  "Mutex wait time",
	Type:   "bar",
	Events: "lastgc",
	Layout: ScatterLayout{
		Yaxis: ScatterYAxis{
			Title:      "seconds / second",
			TickSuffix: "s",
		},
	},
	Subplots: []Subplot{
		{
			Name:    "mutex wait",
			Unitfmt: "%{y:.4s}s",
			Type:    "bar",
		},
	},

	InfoText: `Cumulative metrics are converted to rates by Statsviz so as to be more easily comparable and readable.
<i>mutex wait</i> is <b>/sync/mutex/wait/total</b>, approximate cumulative time goroutines have spent blocked on a sync.Mutex or sync.RWMutex.

This metric is useful for identifying global changes in lock contention. Collect a mutex or block profile using the runtime/pprof package for more detailed contention data.`,
}

var gcScanLayout = Scatter{
	Name:   "TODO(set later)",
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
}

var allocFreeRatesLayout = Scatter{
	Name:  "heap alloc/free rates",
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
<i>Allocations per second</i> is derived by differencing the cumulative <b>/gc/heap/allocs:objects</b> metric.
<i>Frees per second</i> is similarly derived from <b>/gc/heap/frees:objects</b>.`,
}
