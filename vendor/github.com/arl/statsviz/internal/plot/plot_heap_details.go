package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/memory/classes/heap/objects:bytes",
		"/memory/classes/heap/unused:bytes",
		"/memory/classes/heap/free:bytes",
		"/memory/classes/heap/released:bytes",
		"/memory/classes/heap/stacks:bytes",
		"/gc/heap/goal:bytes",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			heapObjects := samples[idx_memory_classes_heap_objects_bytes].Value.Uint64()
			heapUnused := samples[idx_memory_classes_heap_unused_bytes].Value.Uint64()
			heapFree := samples[idx_memory_classes_heap_free_bytes].Value.Uint64()
			heapReleased := samples[idx_memory_classes_heap_released_bytes].Value.Uint64()
			heapStacks := samples[idx_memory_classes_heap_stacks_bytes].Value.Uint64()
			nextGC := samples[idx_gc_heap_goal_bytes].Value.Uint64()

			heapIdle := heapReleased + heapFree
			heapInUse := heapObjects + heapUnused
			heapSys := heapInUse + heapIdle

			return []uint64{
				heapSys,
				heapObjects,
				heapStacks,
				nextGC,
			}
		}
	},

	layout: Scatter{
		Name:   "heap (details)",
		Tags:   []tag{tagGC},
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
			{Unitfmt: "%{y:.4s}B", Name: "heap sys"},
			{Unitfmt: "%{y:.4s}B", Name: "heap objects"},
			{Unitfmt: "%{y:.4s}B", Name: "heap stacks"},
			{Unitfmt: "%{y:.4s}B", Name: "heap goal"},
		},
		InfoText: `
<i>Heap</i> sys is <b>/memory/classes/heap/{objects + unused + released + free}</b>. It's an estimate of all the heap memory obtained from the OS.
<i>Heap objects</i> is <b>/memory/classes/heap/objects</b>, the memory occupied by live objects and dead objects that have not yet been marked free by the GC.
<i>Heap stacks</i> is <b>/memory/classes/heap/stacks</b>, the memory used for stack space.
<i>Heap goal</i> is <b>gc/heap/goal</b>, the heap size target for the end of the GC cycle.`,
	},
})
