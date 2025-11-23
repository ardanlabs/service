package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/memory/classes/metadata/mspan/inuse:bytes",
		"/memory/classes/metadata/mspan/free:bytes",
		"/memory/classes/metadata/mcache/inuse:bytes",
		"/memory/classes/metadata/mcache/free:bytes",
	},
	getvalues: func() getvalues {
		return func(_ time.Time, samples []metrics.Sample) any {
			mspanInUse := samples[idx_memory_classes_metadata_mspan_inuse_bytes].Value.Uint64()
			mspanSys := samples[idx_memory_classes_metadata_mspan_free_bytes].Value.Uint64()
			mcacheInUse := samples[idx_memory_classes_metadata_mcache_inuse_bytes].Value.Uint64()
			mcacheSys := samples[idx_memory_classes_metadata_mcache_free_bytes].Value.Uint64()

			return []uint64{mspanInUse, mspanSys, mcacheInUse, mcacheSys}
		}
	},
	layout: Scatter{
		Name:   "mspan-mcache",
		Tags:   []tag{tagGC},
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
			{Unitfmt: "%{y:.4s}B", Name: "mspan in-use"},
			{Unitfmt: "%{y:.4s}B", Name: "mspan free"},
			{Unitfmt: "%{y:.4s}B", Name: "mcache in-use"},
			{Unitfmt: "%{y:.4s}B", Name: "mcache free"},
		},
		InfoText: `
<i>Mspan in-use</i> is <b>/memory/classes/metadata/mspan/inuse</b>, the memory that is occupied by runtime mspan structures that are currently being used.
<i>Mspan free</i> is <b>/memory/classes/metadata/mspan/free</b>, the memory that is reserved for runtime mspan structures, but not in-use.
<i>Mcache in-use</i> is <b>/memory/classes/metadata/mcache/inuse</b>, the memory that is occupied by runtime mcache structures that are currently being used.
<i>Mcache free</i> is <b>/memory/classes/metadata/mcache/free</b>, the memory that is reserved for runtime mcache structures, but not in-use.
`,
	},
})
