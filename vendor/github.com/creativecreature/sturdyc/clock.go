package sturdyc

import (
	"sync"
	"sync/atomic"
	"time"
)

type Clock interface {
	Now() time.Time
	NewTicker(d time.Duration) (<-chan time.Time, func())
	NewTimer(d time.Duration) (<-chan time.Time, func() bool)
}

type RealClock struct{}

func NewClock() *RealClock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}

func (c *RealClock) NewTicker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}

func (c *RealClock) NewTimer(d time.Duration) (<-chan time.Time, func() bool) {
	t := time.NewTimer(d)
	return t.C, t.Stop
}

type testTimer struct {
	deadline time.Time
	ch       chan time.Time
	stopped  *atomic.Bool
}

type testTicker struct {
	nextTick time.Time
	interval time.Duration
	ch       chan time.Time
	stopped  *atomic.Bool
}

type TestClock struct {
	mu      sync.Mutex
	time    time.Time
	timers  []*testTimer
	tickers []*testTicker
}

func NewTestClock(time time.Time) *TestClock {
	var c TestClock
	c.time = time
	c.timers = make([]*testTimer, 0)
	c.tickers = make([]*testTicker, 0)
	return &c
}

func (c *TestClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if t.Before(c.time) {
		panic("can't go back in time")
	}

	c.time = t
	for _, ticker := range c.tickers {
		if !ticker.stopped.Load() && !ticker.nextTick.Add(ticker.interval).After(c.time) {
			//nolint: durationcheck // This is a test clock, we don't care about overflows.
			nextTick := (c.time.Sub(ticker.nextTick) / ticker.interval) * ticker.interval
			ticker.nextTick = ticker.nextTick.Add(nextTick)
			select {
			case ticker.ch <- c.time:
			default:
			}
		}
	}

	unfiredTimers := make([]*testTimer, 0)
	for i, timer := range c.timers {
		if timer.deadline.After(c.time) && !timer.stopped.Load() {
			unfiredTimers = append(unfiredTimers, c.timers[i])
			continue
		}
		timer.stopped.Store(true)
		timer.ch <- c.time
	}
	c.timers = unfiredTimers
}

func (c *TestClock) Add(d time.Duration) {
	c.Set(c.time.Add(d))
}

func (c *TestClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.time
}

func (c *TestClock) NewTicker(d time.Duration) (<-chan time.Time, func()) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan time.Time, 1)
	stopped := &atomic.Bool{}
	ticker := &testTicker{nextTick: c.time, interval: d, ch: ch, stopped: stopped}
	c.tickers = append(c.tickers, ticker)
	stop := func() {
		stopped.Store(true)
	}

	return ch, stop
}

func (c *TestClock) NewTimer(d time.Duration) (<-chan time.Time, func() bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan time.Time, 1)
	stopped := &atomic.Bool{}

	// Fire the timer straight away if the duration is less than zero.
	if d <= 0 {
		ch <- c.time
		return ch, func() bool { return false }
	}

	timer := &testTimer{deadline: c.time.Add(d), ch: ch, stopped: stopped}
	c.timers = append(c.timers, timer)
	stop := func() bool {
		return stopped.CompareAndSwap(false, true)
	}

	return ch, stop
}
