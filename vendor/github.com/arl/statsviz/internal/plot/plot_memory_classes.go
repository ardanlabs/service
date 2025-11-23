package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/memory/classes/os-stacks:bytes",
		"/memory/classes/other:bytes",
		"/memory/classes/profiling/buckets:bytes",
		"/memory/classes/total:bytes",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			osStacks := samples[idx_memory_classes_os_stacks_bytes].Value.Uint64()
			other := samples[idx_memory_classes_other_bytes].Value.Uint64()
			profBuckets := samples[idx_memory_classes_profiling_buckets_bytes].Value.Uint64()
			total := samples[idx_memory_classes_total_bytes].Value.Uint64()

			return []uint64{
				osStacks,
				other,
				profBuckets,
				total,
			}
		}
	},
	layout: Scatter{
		Name:   "memory-classes",
		Tags:   []tag{tagGC},
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
			{Unitfmt: "%{y:.4s}B", Name: "os stacks"},
			{Unitfmt: "%{y:.4s}B", Name: "other"},
			{Unitfmt: "%{y:.4s}B", Name: "profiling buckets"},
			{Unitfmt: "%{y:.4s}B", Name: "total"},
		},

		InfoText: `
<i>OS stacks</i> is <b>/memory/classes/os-stacks</b>, stack memory allocated by the underlying operating system.
<i>Other</i> is <b>/memory/classes/other</b>, memory used by execution trace buffers, structures for debugging the runtime, finalizer and profiler specials, and more.
<i>Profiling buckets</i> is <b>/memory/classes/profiling/buckets</b>, memory that is used by the stack trace hash map used for profiling.
<i>Total</i> is <b>/memory/classes/total</b>, all memory mapped by the Go runtime into the current process as read-write.`,
	},
})
