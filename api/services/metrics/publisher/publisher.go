// Package publisher manages the publishing of metrics.
package publisher

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ardanlabs/service/foundation/logger"
)

// Set of possible publisher types.
const (
	TypeStdout  = "stdout"
	TypeDatadog = "datadog"
)

// Collector defines a contract a collector must support
// so a consumer can retrieve metrics.
type Collector interface {
	Collect() (map[string]any, error)
}

// Publisher defines a handler function that will be called
// on each interval.
type Publisher func(map[string]any)

// Publish provides the ability to receive metrics
// on an interval.
type Publish struct {
	log       *logger.Logger
	collector Collector
	publisher []Publisher
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

// New creates a Publish for consuming and publishing metrics.
func New(log *logger.Logger, collector Collector, interval time.Duration, publisher ...Publisher) (*Publish, error) {
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
		p.log.Error(context.Background(), "publish", "status", "collect data", "err", err)
		return
	}

	for _, pub := range p.publisher {
		pub(data)
	}
}

// Stdout provide our basic publishing.
type Stdout struct {
	log *logger.Logger
}

// NewStdout initializes stdout for publishing metrics.
func NewStdout(log *logger.Logger) *Stdout {
	return &Stdout{log}
}

// Publish publishers for writing to stdout.
func (s *Stdout) Publish(data map[string]any) {
	ctx := context.Background()

	rawJSON, err := json.Marshal(data)
	if err != nil {
		s.log.Error(ctx, "stdout", "status", "marshal data", "err", err)
		return
	}

	var d map[string]any
	if err := json.Unmarshal(rawJSON, &d); err != nil {
		s.log.Error(ctx, "stdout", "status", "unmarshal data", "err", err)
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

	out, err := json.Marshal(d)
	if err != nil {
		return
	}
	s.log.Info(ctx, "stdout", "data", string(out))
}
