package sturdyc

import (
	"sync"
	"sync/atomic"
	"time"
)

// Clock is an abstraction for time.Time package that allows for testing.
type Clock interface {
	Now() time.Time
	NewTicker(d time.Duration) (<-chan time.Time, func())
	NewTimer(d time.Duration) (<-chan time.Time, func() bool)
	Since(t time.Time) time.Duration
}

// RealClock provides functions that wraps the real time.Time package.
type RealClock struct{}

// NewClock returns a new RealClock.
func NewClock() *RealClock {
	return &RealClock{}
}

// Now wraps time.Now() from the standard library.
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// NewTicker returns the channel and stop function from the ticker from the standard library.
func (c *RealClock) NewTicker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}

// NewTimer returns the channel and stop function from the timer from the standard library.
func (c *RealClock) NewTimer(d time.Duration) (<-chan time.Time, func() bool) {
	t := time.NewTimer(d)
	return t.C, t.Stop
}

// Since wraps time.Since() from the standard library.
func (c *RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
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

// TestClock is a clock that satisfies the Clock interface. It should only be used for testing.
type TestClock struct {
	mu      sync.Mutex
	time    time.Time
	timers  []*testTimer
	tickers []*testTicker
}

// NewTestClock returns a new TestClock with the specified time.
func NewTestClock(time time.Time) *TestClock {
	var c TestClock
	c.time = time
	c.timers = make([]*testTimer, 0)
	c.tickers = make([]*testTicker, 0)
	return &c
}

// Set sets the internal time of the test clock and triggers any timers or tickers that should fire.
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

// Add adds the duration to the internal time of the test clock
// and triggers any timers or tickers that should fire.
func (c *TestClock) Add(d time.Duration) {
	c.Set(c.time.Add(d))
}

// Now returns the internal time of the test clock.
func (c *TestClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.time
}

// NewTicker creates a new ticker that will fire every time
// the internal clock advances by the specified duration.
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

// NewTimer creates a new timer that will fire once the internal time
// of the clock has been advanced passed the specified duration.
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

// Since returns the duration between the internal time of the clock and the specified time.
func (c *TestClock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}
