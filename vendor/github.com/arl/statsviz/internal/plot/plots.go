package plot

import (
	"math"
	"runtime/metrics"
	"time"
)

type plotDesc struct {
	name    string
	tags    []string
	metrics []string
	layout  any

	// make creates the state (support struct) for the plot.
	make func(indices ...int) metricsGetter
}

var (
	plotDescs []plotDesc

	metricDescs = metrics.All()
	metricIdx   map[string]int
)

func init() {
	// We need a first set of sample in order to dimension and process the
	// heatmaps buckets.
	samples := make([]metrics.Sample, len(metricDescs))
	metricIdx = make(map[string]int)

	for i := range samples {
		samples[i].Name = metricDescs[i].Name
		metricIdx[samples[i].Name] = i
	}
	metrics.Read(samples)

	plotDescs = []plotDesc{
		{
			name: "garbage collection",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/gomemlimit:bytes",
				"/gc/heap/live:bytes",
				"/gc/heap/goal:bytes",
				"/memory/classes/total:bytes",
				"/memory/classes/heap/released:bytes",
			},
			layout: garbageCollectionLayout,
			make:   makeGarbageCollection,
		},
		{
			name: "heap (details)",
			tags: []string{"gc"},
			metrics: []string{
				"/memory/classes/heap/objects:bytes",
				"/memory/classes/heap/unused:bytes",
				"/memory/classes/heap/free:bytes",
				"/memory/classes/heap/released:bytes",
				"/memory/classes/heap/stacks:bytes",
				"/gc/heap/goal:bytes",
			},
			layout: heapDetailslLayout,
			make:   makeHeapDetails,
		},
		{
			name: "live-objects",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/heap/objects:objects",
			},
			layout: liveObjectsLayout,
			make:   makeLiveObjects,
		},
		{
			name: "live-bytes",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/heap/allocs:bytes",
				"/gc/heap/frees:bytes",
			},
			layout: liveBytesLayout,
			make:   makeLiveBytes,
		},
		{
			name: "mspan-mcache",
			tags: []string{"gc"},
			metrics: []string{
				"/memory/classes/metadata/mspan/inuse:bytes",
				"/memory/classes/metadata/mspan/free:bytes",
				"/memory/classes/metadata/mcache/inuse:bytes",
				"/memory/classes/metadata/mcache/free:bytes",
			},
			layout: mspanMCacheLayout,
			make:   makeMSpanMCache,
		},
		{
			name: "goroutines",
			tags: []string{"scheduler"},
			metrics: []string{
				"/sched/goroutines:goroutines",
			},
			layout: goroutinesLayout,
			make:   makeGoroutines,
		},
		{
			name: "size-classes",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/heap/allocs-by-size:bytes",
				"/gc/heap/frees-by-size:bytes",
			},
			layout: sizeClassesLayout(samples),
			make:   makeSizeClasses,
		},
		{
			name: "gc-pauses",
			tags: []string{"scheduler"},
			metrics: []string{
				"/sched/pauses/total/gc:seconds",
			},
			layout: gcPausesLayout(samples),
			make:   makeGCPauses,
		},
		{
			name: "runnable-time",
			tags: []string{"scheduler"},
			metrics: []string{
				"/sched/latencies:seconds",
			},
			layout: runnableTimeLayout(samples),
			make:   makeRunnableTime,
		},
		{
			name: "sched-events",
			tags: []string{"scheduler"},
			metrics: []string{
				"/sched/latencies:seconds",
				"/sched/gomaxprocs:threads",
			},
			layout: schedEventsLayout,
			make:   makeSchedEvents,
		},
		{
			name: "cgo",
			tags: []string{"misc"},
			metrics: []string{
				"/cgo/go-to-c-calls:calls",
			},
			layout: cgoLayout,
			make:   makeCGO,
		},
		{
			name: "gc-stack-size",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/stack/starting-size:bytes",
			},
			layout: gcStackSizeLayout,
			make:   makeGCStackSize,
		},
		{
			name: "gc-cycles",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/cycles/automatic:gc-cycles",
				"/gc/cycles/forced:gc-cycles",
				"/gc/cycles/total:gc-cycles",
			},
			layout: gcCyclesLayout,
			make:   makeGCCycles,
		},
		{
			name: "memory-classes",
			tags: []string{"gc"},
			metrics: []string{
				"/memory/classes/os-stacks:bytes",
				"/memory/classes/other:bytes",
				"/memory/classes/profiling/buckets:bytes",
				"/memory/classes/total:bytes",
			},
			layout: memoryClassesLayout,
			make:   makeMemoryClasses,
		},
		{
			name: "cpu-gc",
			tags: []string{"cpu", "gc"},
			metrics: []string{
				"/cpu/classes/gc/mark/assist:cpu-seconds",
				"/cpu/classes/gc/mark/dedicated:cpu-seconds",
				"/cpu/classes/gc/mark/idle:cpu-seconds",
				"/cpu/classes/gc/pause:cpu-seconds",
			},
			layout: cpuGCLayout,
			make:   makeCPUgc,
		},
		{
			name: "cpu-scavenger",
			tags: []string{"cpu", "gc"},
			metrics: []string{
				"/cpu/classes/scavenge/assist:cpu-seconds",
				"/cpu/classes/scavenge/background:cpu-seconds",
			},
			layout: cpuScavengerLayout,
			make:   makeCPUscavenger,
		},
		{
			name: "cpu-overall",
			tags: []string{"cpu"},
			metrics: []string{
				"/cpu/classes/user:cpu-seconds",
				"/cpu/classes/scavenge/total:cpu-seconds",
				"/cpu/classes/idle:cpu-seconds",
				"/cpu/classes/gc/total:cpu-seconds",
				"/cpu/classes/total:cpu-seconds",
			},
			layout: cpuOverallLayout,
			make:   makeCPUoverall,
		},
		{
			name: "mutex-wait",
			tags: []string{"misc"},
			metrics: []string{
				"/sync/mutex/wait/total:seconds",
			},
			layout: mutexWaitLayout,
			make:   makeMutexWait,
		},
		{
			name: "gc-scan",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/scan/globals:bytes",
				"/gc/scan/heap:bytes",
				"/gc/scan/stack:bytes",
			},
			layout: gcScanLayout,
			make:   makeGCScan,
		},
		{
			name: "alloc-free-rate",
			tags: []string{"gc"},
			metrics: []string{
				"/gc/heap/allocs:objects",
				"/gc/heap/frees:objects",
			},
			layout: allocFreeRatesLayout,
			make:   makeAllocFreeRates,
		},
	}
}

// garbage collection
type garbageCollection struct {
	idxmemlimit     int
	idxheaplive     int
	idxheapgoal     int
	idxmemtotal     int
	idxheapreleased int
}

func makeGarbageCollection(indices ...int) metricsGetter {
	return &garbageCollection{
		idxmemlimit:     indices[0],
		idxheaplive:     indices[1],
		idxheapgoal:     indices[2],
		idxmemtotal:     indices[3],
		idxheapreleased: indices[4],
	}
}

func (p *garbageCollection) values(samples []metrics.Sample) any {
	memLimit := samples[p.idxmemlimit].Value.Uint64()
	heapLive := samples[p.idxheaplive].Value.Uint64()
	heapGoal := samples[p.idxheapgoal].Value.Uint64()
	memTotal := samples[p.idxmemtotal].Value.Uint64()
	heapReleased := samples[p.idxheapreleased].Value.Uint64()

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

// heap (details)

type heapDetails struct {
	idxobj      int
	idxunused   int
	idxfree     int
	idxreleased int
	idxstacks   int
	idxgoal     int
}

func makeHeapDetails(indices ...int) metricsGetter {
	return &heapDetails{
		idxobj:      indices[0],
		idxunused:   indices[1],
		idxfree:     indices[2],
		idxreleased: indices[3],
		idxstacks:   indices[4],
		idxgoal:     indices[5],
	}
}

func (p *heapDetails) values(samples []metrics.Sample) any {
	heapObjects := samples[p.idxobj].Value.Uint64()
	heapUnused := samples[p.idxunused].Value.Uint64()
	heapFree := samples[p.idxfree].Value.Uint64()
	heapReleased := samples[p.idxreleased].Value.Uint64()
	heapStacks := samples[p.idxstacks].Value.Uint64()
	nextGC := samples[p.idxgoal].Value.Uint64()

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

// live objects

type liveObjects struct {
	idxobjects int
}

func makeLiveObjects(indices ...int) metricsGetter {
	return &liveObjects{
		idxobjects: indices[0],
	}
}

func (p *liveObjects) values(samples []metrics.Sample) any {
	gcHeapObjects := samples[p.idxobjects].Value.Uint64()
	return []uint64{
		gcHeapObjects,
	}
}

// live bytes

type liveBytes struct {
	idxallocs int
	idxfrees  int
}

func makeLiveBytes(indices ...int) metricsGetter {
	return &liveBytes{
		idxallocs: indices[0],
		idxfrees:  indices[1],
	}
}

func (p *liveBytes) values(samples []metrics.Sample) any {
	allocBytes := samples[p.idxallocs].Value.Uint64()
	freedBytes := samples[p.idxfrees].Value.Uint64()
	return []uint64{
		allocBytes - freedBytes,
	}
}

// mspan mcache

type mspanMcache struct {
	enabled bool

	idxmspanInuse  int
	idxmspanFree   int
	idxmcacheInuse int
	idxmcacheFree  int
}

func makeMSpanMCache(indices ...int) metricsGetter {
	return &mspanMcache{
		idxmspanInuse:  indices[0],
		idxmspanFree:   indices[1],
		idxmcacheInuse: indices[2],
		idxmcacheFree:  indices[3],
	}
}

func (p *mspanMcache) values(samples []metrics.Sample) any {
	mspanInUse := samples[p.idxmspanInuse].Value.Uint64()
	mspanSys := samples[p.idxmspanFree].Value.Uint64()
	mcacheInUse := samples[p.idxmcacheInuse].Value.Uint64()
	mcacheSys := samples[p.idxmcacheFree].Value.Uint64()
	return []uint64{
		mspanInUse,
		mspanSys,
		mcacheInUse,
		mcacheSys,
	}
}

// goroutines

type goroutines struct {
	idxgs int
}

func makeGoroutines(indices ...int) metricsGetter {
	return &goroutines{
		idxgs: indices[0],
	}
}

func (p *goroutines) values(samples []metrics.Sample) any {
	return []uint64{samples[p.idxgs].Value.Uint64()}
}

// size classes

type sizeClasses struct {
	sizeClasses []uint64

	idxallocs int
	idxfrees  int
}

func makeSizeClasses(indices ...int) metricsGetter {
	return &sizeClasses{
		idxallocs: indices[0],
		idxfrees:  indices[1],
	}
}

func (p *sizeClasses) values(samples []metrics.Sample) any {
	allocsBySize := samples[p.idxallocs].Value.Float64Histogram()
	freesBySize := samples[p.idxfrees].Value.Float64Histogram()

	if p.sizeClasses == nil {
		p.sizeClasses = make([]uint64, len(allocsBySize.Counts))
	}

	for i := range p.sizeClasses {
		p.sizeClasses[i] = allocsBySize.Counts[i] - freesBySize.Counts[i]
	}
	return p.sizeClasses
}

// gc pauses

type gcpauses struct {
	histfactor int
	counts     [maxBuckets]uint64

	idxgcpauses int
}

func makeGCPauses(indices ...int) metricsGetter {
	return &gcpauses{
		idxgcpauses: indices[0],
	}
}

func (p *gcpauses) values(samples []metrics.Sample) any {
	if p.histfactor == 0 {
		gcpauses := samples[p.idxgcpauses].Value.Float64Histogram()
		p.histfactor = downsampleFactor(len(gcpauses.Buckets), maxBuckets)
	}

	gcpauses := samples[p.idxgcpauses].Value.Float64Histogram()
	return downsampleCounts(gcpauses, p.histfactor, p.counts[:])
}

// runnable time

type runnableTime struct {
	histfactor int
	counts     [maxBuckets]uint64

	idxschedlat int
}

func makeRunnableTime(indices ...int) metricsGetter {
	return &runnableTime{
		idxschedlat: indices[0],
	}
}

func (p *runnableTime) values(samples []metrics.Sample) any {
	if p.histfactor == 0 {
		schedlat := samples[p.idxschedlat].Value.Float64Histogram()
		p.histfactor = downsampleFactor(len(schedlat.Buckets), maxBuckets)
	}

	schedlat := samples[p.idxschedlat].Value.Float64Histogram()

	return downsampleCounts(schedlat, p.histfactor, p.counts[:])
}

// sched events

type schedEvents struct {
	idxschedlat   int
	idxGomaxprocs int
	lasttot       uint64
}

func makeSchedEvents(indices ...int) metricsGetter {
	return &schedEvents{
		idxschedlat:   indices[0],
		idxGomaxprocs: indices[1],
		lasttot:       math.MaxUint64,
	}
}

// gTrackingPeriod is currently always 8. Guard it behind build tags when that
// changes. See https://github.com/golang/go/blob/go1.18.4/src/runtime/runtime2.go#L502-L504
const currentGtrackingPeriod = 8

// TODO show scheduling events per seconds
func (p *schedEvents) values(samples []metrics.Sample) any {
	schedlat := samples[p.idxschedlat].Value.Float64Histogram()
	gomaxprocs := samples[p.idxGomaxprocs].Value.Uint64()

	total := uint64(0)
	for _, v := range schedlat.Counts {
		total += v
	}
	total *= currentGtrackingPeriod

	curtot := total - p.lasttot
	if p.lasttot == math.MaxUint64 {
		// We don't want a big spike at statsviz launch in case the process has
		// been running for some time and curtot is high.
		curtot = 0
	}
	p.lasttot = total

	ftot := float64(curtot)

	return []float64{
		ftot,
		ftot / float64(gomaxprocs),
	}
}

// cgo

type cgo struct {
	idxgo2c  int
	lastgo2c uint64
}

func makeCGO(indices ...int) metricsGetter {
	return &cgo{
		idxgo2c:  indices[0],
		lastgo2c: math.MaxUint64,
	}
}

// TODO show cgo calls per second
func (p *cgo) values(samples []metrics.Sample) any {
	go2c := samples[p.idxgo2c].Value.Uint64()
	curgo2c := go2c - p.lastgo2c
	if p.lastgo2c == math.MaxUint64 {
		curgo2c = 0
	}
	p.lastgo2c = go2c

	return []uint64{curgo2c}
}

// gc stack size

type gcStackSize struct {
	idxstack int
}

func makeGCStackSize(indices ...int) metricsGetter {
	return &gcStackSize{
		idxstack: indices[0],
	}
}

func (p *gcStackSize) values(samples []metrics.Sample) any {
	stackSize := samples[p.idxstack].Value.Uint64()
	return []uint64{stackSize}
}

// gc cycles

type gcCycles struct {
	idxAutomatic int
	idxForced    int
	idxTotal     int

	lastAuto, lastForced, lastTotal uint64
}

func makeGCCycles(indices ...int) metricsGetter {
	return &gcCycles{
		idxAutomatic: indices[0],
		idxForced:    indices[1],
		idxTotal:     indices[2],
	}
}

func (p *gcCycles) values(samples []metrics.Sample) any {
	total := samples[p.idxTotal].Value.Uint64()
	auto := samples[p.idxAutomatic].Value.Uint64()
	forced := samples[p.idxForced].Value.Uint64()

	if p.lastTotal == 0 {
		p.lastTotal = total
		p.lastForced = forced
		p.lastAuto = auto
		return []uint64{0, 0}
	}

	ret := []uint64{
		auto - p.lastAuto,
		forced - p.lastForced,
	}

	p.lastForced = forced
	p.lastAuto = auto

	return ret
}

// memory classes

type memoryClasses struct {
	idxOSStacks    int
	idxOther       int
	idxProfBuckets int
	idxTotal       int
}

func makeMemoryClasses(indices ...int) metricsGetter {
	return &memoryClasses{
		idxOSStacks:    indices[0],
		idxOther:       indices[1],
		idxProfBuckets: indices[2],
		idxTotal:       indices[3],
	}
}

func (p *memoryClasses) values(samples []metrics.Sample) any {
	osStacks := samples[p.idxOSStacks].Value.Uint64()
	other := samples[p.idxOther].Value.Uint64()
	profBuckets := samples[p.idxProfBuckets].Value.Uint64()
	total := samples[p.idxTotal].Value.Uint64()

	return []uint64{
		osStacks,
		other,
		profBuckets,
		total,
	}
}

// cpu (gc)

type CPUgc struct {
	idxMarkAssist    int
	idxMarkDedicated int
	idxMarkIdle      int
	idxPause         int
	idxTotal         int

	lastTime time.Time

	lastMarkAssist    float64
	lastMarkDedicated float64
	lastMarkIdle      float64
	lastPause         float64
}

func makeCPUgc(indices ...int) metricsGetter {
	return &CPUgc{
		idxMarkAssist:    indices[0],
		idxMarkDedicated: indices[1],
		idxMarkIdle:      indices[2],
		idxPause:         indices[3],
	}
}

func (p *CPUgc) values(samples []metrics.Sample) any {
	curMarkAssist := samples[p.idxMarkAssist].Value.Float64()
	curMarkDedicated := samples[p.idxMarkDedicated].Value.Float64()
	curMarkIdle := samples[p.idxMarkIdle].Value.Float64()
	curPause := samples[p.idxPause].Value.Float64()

	if p.lastTime.IsZero() {
		p.lastMarkAssist = curMarkAssist
		p.lastMarkDedicated = curMarkDedicated
		p.lastMarkIdle = curMarkIdle
		p.lastPause = curPause
		p.lastTime = time.Now()

		return []float64{0, 0, 0, 0, 0}
	}

	t := time.Since(p.lastTime).Seconds()

	markAssist := (curMarkAssist - p.lastMarkAssist) / t
	markDedicated := (curMarkDedicated - p.lastMarkDedicated) / t
	markIdle := (curMarkIdle - p.lastMarkIdle) / t
	pause := (curPause - p.lastPause) / t

	p.lastMarkAssist = curMarkAssist
	p.lastMarkDedicated = curMarkDedicated
	p.lastMarkIdle = curMarkIdle
	p.lastPause = curPause
	p.lastTime = time.Now()

	return []float64{
		markAssist,
		markDedicated,
		markIdle,
		pause,
	}
}

// cpu (scavenger)

type cpuScavenger struct {
	idxScavengeAssist     int
	idxScavengeBackground int

	lastTime time.Time

	lastScavengeAssist     float64
	lastScavengeBackground float64
}

func makeCPUscavenger(indices ...int) metricsGetter {
	return &cpuScavenger{
		idxScavengeAssist:     indices[0],
		idxScavengeBackground: indices[1],
	}
}

func (p *cpuScavenger) values(samples []metrics.Sample) any {
	curScavengeAssist := samples[p.idxScavengeAssist].Value.Float64()
	curScavengeBackground := samples[p.idxScavengeBackground].Value.Float64()

	if p.lastTime.IsZero() {
		p.lastScavengeAssist = curScavengeAssist
		p.lastScavengeBackground = curScavengeBackground
		p.lastTime = time.Now()

		return []float64{0, 0, 0, 0, 0}
	}

	t := time.Since(p.lastTime).Seconds()

	scavengeAssist := (curScavengeAssist - p.lastScavengeAssist) / t
	scavengeBackground := (curScavengeBackground - p.lastScavengeBackground) / t

	p.lastScavengeAssist = curScavengeAssist
	p.lastScavengeBackground = curScavengeBackground

	return []float64{
		scavengeAssist,
		scavengeBackground,
	}
}

// cpu overall

type cpuOverall struct {
	idxUser     int
	idxScavenge int
	idxIdle     int
	idxGCtotal  int
	idxTotal    int

	lastTime     time.Time
	lastUser     float64
	lastScavenge float64
	lastIdle     float64
	lastGCtotal  float64
	lastTotal    float64
}

func makeCPUoverall(indices ...int) metricsGetter {
	return &cpuOverall{
		idxUser:     indices[0],
		idxScavenge: indices[1],
		idxIdle:     indices[2],
		idxGCtotal:  indices[3],
		idxTotal:    indices[4],
	}
}

func (p *cpuOverall) values(samples []metrics.Sample) any {
	curUser := samples[p.idxUser].Value.Float64()
	curScavenge := samples[p.idxScavenge].Value.Float64()
	curIdle := samples[p.idxIdle].Value.Float64()
	curGCtotal := samples[p.idxGCtotal].Value.Float64()
	curTotal := samples[p.idxTotal].Value.Float64()

	if p.lastTime.IsZero() {
		p.lastUser = curUser
		p.lastScavenge = curScavenge
		p.lastIdle = curIdle
		p.lastGCtotal = curGCtotal
		p.lastTotal = curTotal

		p.lastTime = time.Now()
		return []float64{0, 0, 0, 0, 0}
	}

	t := time.Since(p.lastTime).Seconds()

	user := (curUser - p.lastUser) / t
	scavenge := (curScavenge - p.lastScavenge) / t
	idle := (curIdle - p.lastIdle) / t
	gcTotal := (curGCtotal - p.lastGCtotal) / t
	total := (curTotal - p.lastTotal) / t

	p.lastUser = curUser
	p.lastScavenge = curScavenge
	p.lastIdle = curIdle
	p.lastGCtotal = curGCtotal
	p.lastTotal = curTotal

	return []float64{
		user,
		scavenge,
		idle,
		gcTotal,
		total,
	}
}

// mutex wait

type mutexWait struct {
	idxMutexWait int

	lastTime      time.Time
	lastMutexWait float64
}

func makeMutexWait(indices ...int) metricsGetter {
	return &mutexWait{
		idxMutexWait: indices[0],
	}
}

func (p *mutexWait) values(samples []metrics.Sample) any {
	if p.lastTime.IsZero() {
		p.lastTime = time.Now()
		p.lastMutexWait = samples[p.idxMutexWait].Value.Float64()

		return []float64{0}
	}

	t := time.Since(p.lastTime).Seconds()

	mutexWait := (samples[p.idxMutexWait].Value.Float64() - p.lastMutexWait) / t

	p.lastMutexWait = samples[p.idxMutexWait].Value.Float64()
	p.lastTime = time.Now()

	return []float64{
		mutexWait,
	}
}

// gc scan

type gcScan struct {
	idxGlobals int
	idxHeap    int
	idxStack   int
}

func makeGCScan(indices ...int) metricsGetter {
	return &gcScan{
		idxGlobals: indices[0],
		idxHeap:    indices[1],
		idxStack:   indices[2],
	}
}

func (p *gcScan) values(samples []metrics.Sample) any {
	globals := samples[p.idxGlobals].Value.Uint64()
	heap := samples[p.idxHeap].Value.Uint64()
	stack := samples[p.idxStack].Value.Uint64()
	return []uint64{
		globals,
		heap,
		stack,
	}
}

// alloc/free rates
type allocFreeRates struct {
	idxallocs int
	idxfrees  int

	lasttime   time.Time
	lastallocs uint64
	lastfrees  uint64
}

func makeAllocFreeRates(indices ...int) metricsGetter {
	return &allocFreeRates{
		idxallocs: indices[0],
		idxfrees:  indices[1],
	}
}

func (p *allocFreeRates) values(samples []metrics.Sample) any {
	if p.lasttime.IsZero() {
		p.lasttime = time.Now()
		p.lastallocs = samples[p.idxallocs].Value.Uint64()
		p.lastfrees = samples[p.idxfrees].Value.Uint64()

		return []float64{0, 0}
	}

	t := time.Since(p.lasttime).Seconds()

	allocs := float64(samples[p.idxallocs].Value.Uint64()-p.lastallocs) / t
	frees := float64(samples[p.idxfrees].Value.Uint64()-p.lastfrees) / t

	p.lastallocs = samples[p.idxallocs].Value.Uint64()
	p.lastfrees = samples[p.idxfrees].Value.Uint64()
	p.lasttime = time.Now()

	return []float64{
		allocs,
		frees,
	}
}

/*
 * helpers
 */

func floatseq(n int) []float64 {
	seq := make([]float64, n)
	for i := range n {
		seq[i] = float64(i)
	}
	return seq
}
