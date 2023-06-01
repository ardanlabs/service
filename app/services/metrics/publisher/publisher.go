// Package publisher manages the publishing of metrics.
package publisher

import (
	"encoding/json"
	"sync"
	"time"

	"golang.org/x/exp/slog"
)

// Set of possible publisher types.
const (
	TypeStdout  = "stdout"
	TypeDatadog = "datadog"
)

// =============================================================================

// Collector defines a contract a collector must support
// so a consumer can retrieve metrics.
type Collector interface {
	Collect() (map[string]any, error)
}

// =============================================================================

// Publisher defines a handler function that will be called
// on each interval.
type Publisher func(map[string]any)

// Publish provides the ability to receive metrics
// on an interval.
type Publish struct {
	log       *slog.Logger
	collector Collector
	publisher []Publisher
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

// New creates a Publish for consuming and publishing metrics.
func New(log *slog.Logger, collector Collector, interval time.Duration, publisher ...Publisher) (*Publish, error) {
	p := Publish{
		log:       log,
		collector: collector,
		publisher: publisher,
		timer:     time.NewTimer(interval),
		shutdown:  make(chan struct{}),
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			p.timer.Reset(interval)
			select {
			case <-p.timer.C:
				p.update()
			case <-p.shutdown:
				return
			}
		}
	}()

	return &p, nil
}

// Stop is used to shut down the goroutine collecting metrics.
func (p *Publish) Stop() {
	close(p.shutdown)
	p.wg.Wait()
}

// update pulls the metrics and publishes them to the specified system.
func (p *Publish) update() {
	data, err := p.collector.Collect()
	if err != nil {
		p.log.Info("publish", "status", "collect data", "ERROR", err)
		return
	}

	for _, pub := range p.publisher {
		pub(data)
	}
}

// =============================================================================

// Stdout provide our basic publishing.
type Stdout struct {
	log *slog.Logger
}

// NewStdout initializes stdout for publishing metrics.
func NewStdout(log *slog.Logger) *Stdout {
	return &Stdout{log}
}

// Publish publishers for writing to stdout.
func (s *Stdout) Publish(data map[string]any) {
	rawJSON, err := json.Marshal(data)
	if err != nil {
		s.log.Info("stdout", "status", "marshal data", "ERROR", err)
		return
	}

	var d map[string]any
	if err := json.Unmarshal(rawJSON, &d); err != nil {
		s.log.Info("stdout", "status", "unmarshal data", "ERROR", err)
		return
	}

	// Add heap value into the data set.
	memStats, ok := (d["memstats"]).(map[string]any)
	if ok {
		d["heap"] = memStats["Alloc"]
	}

	// Remove unnecessary keys.
	delete(d, "memstats")
	delete(d, "cmdline")

	out, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		return
	}
	s.log.Info("stdout", "data", string(out))
}
