package plot

import (
	"runtime/debug"
	"time"
)

// delta returns a function that computes the delta between successive calls.
func delta[T uint64 | float64]() func(T) T {
	first := true
	var last T
	return func(cur T) T {
		delta := cur - last
		if first {
			delta = 0
			first = false
		}
		last = cur
		return delta
	}
}

// rate returns a function that computes the rate of change per second.
func rate[T uint64 | float64]() func(time.Time, T) float64 {
	var last T
	var lastTime time.Time

	return func(now time.Time, cur T) float64 {
		if lastTime.IsZero() {
			last = cur
			lastTime = now
			return 0
		}

		t := now.Sub(lastTime).Seconds()
		rate := float64(cur-last) / t

		last = cur
		lastTime = now

		return rate
	}
}

func goversion() string {
	bnfo, ok := debug.ReadBuildInfo()
	if ok {
		return bnfo.GoVersion
	}

	return "<unknown version>"
}
