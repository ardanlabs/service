//go:build go1.26

package plot

import (
	"runtime/metrics"
	"time"
)

var _ = register(description{
	metrics: []string{
		"/sched/goroutines:goroutines",
		"/sched/goroutines-created:goroutines",
		"/sched/goroutines/not-in-go:goroutines",
		"/sched/goroutines/runnable:goroutines",
		"/sched/goroutines/running:goroutines",
		"/sched/goroutines/waiting:goroutines",
	},
	getvalues: func() getvalues {
		deltaCreated := delta[uint64]()

		return func(_ time.Time, samples []metrics.Sample) any {
			created := deltaCreated(samples[idx_sched_goroutines_created_goroutines].Value.Uint64())
			goroutines := samples[idx_sched_goroutines_goroutines].Value.Uint64()
			notInGo := samples[idx_sched_goroutines_not_in_go_goroutines].Value.Uint64()
			runnable := samples[idx_sched_goroutines_runnable_goroutines].Value.Uint64()
			running := samples[idx_sched_goroutines_running_goroutines].Value.Uint64()
			waiting := samples[idx_sched_goroutines_waiting_goroutines].Value.Uint64()

			return []uint64{created, goroutines, notInGo, runnable, running, waiting}
		}
	},

	layout: Scatter{
		Name:  "goroutines",
		Tags:  []tag{tagScheduler},
		Title: "Goroutines",
		Type:  "scatter",
		Layout: ScatterLayout{
			Yaxis: ScatterYAxis{
				Title: "goroutines",
			},
		},
		Subplots: []Subplot{
			{Name: "created", Unitfmt: "%{y}", Type: "bar"},
			{Name: "goroutines", Unitfmt: "%{y}"},
			{Name: "not in Go", Unitfmt: "%{y}"},
			{Name: "runnable", Unitfmt: "%{y}"},
			{Name: "running", Unitfmt: "%{y}"},
			{Name: "waiting", Unitfmt: "%{y}"},
		},
		InfoText: `<i>Goroutines</i> is <b>/sched/goroutines</b>, the count of live goroutines.
<i>Created</i> is the delta of <b>/sched/goroutines-created</b>, the cumulative number of created goroutines.
<i>Not in Go</i> is <b>/sched/goroutines/not-in-go</b>, the approximate count of goroutines running or blocked in a system call or cgo call.
<i>Runnable</i> is <b>/sched/goroutines/runnable</b>, the approximate count of goroutines ready to execute, but not executing.
<i>Running</i> is <b>/sched/goroutines/running</b>, the approximate count of goroutines executing.
<i>Waiting</i> is <b>/sched/goroutines/waiting</b>, the approximate count of goroutines waiting on a resource (I/O or sync primitives).`,
	},
})
