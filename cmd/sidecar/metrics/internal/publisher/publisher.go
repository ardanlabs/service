package publisher

import (
	"encoding/json"
	"log"
	"sync"
	"time"
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
	Collect() (map[string]interface{}, error)
}

// =============================================================================

// Publisher defines a handler function that will be called
// on each interval.
type Publisher func(map[string]interface{})

// Publish provides the ability to receive metrics
// on an interval.
type Publish struct {
	log       *log.Logger
	collector Collector
	publisher []Publisher
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

// New creates a Publish for consuming and publishing metrics.
func New(log *log.Logger, collector Collector, interval time.Duration, publisher ...Publisher) (*Publish, error) {
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

// Stop is used to shutdown the goroutine collecting metrics.
func (p *Publish) Stop() {
	close(p.shutdown)
	p.wg.Wait()
}

// update pulls the metrics and publishes them to the specified system.
func (p *Publish) update() {
	data, err := p.collector.Collect()
	if err != nil {
		p.log.Println(err)
		return
	}

	for _, pub := range p.publisher {
		pub(data)
	}
}

// =============================================================================

// Stdout provide our basic publishing.
type Stdout struct {
	log *log.Logger
}

// NewStdout initializes stdout for publishing metrics.
func NewStdout(log *log.Logger) *Stdout {
	return &Stdout{log}
}

// Publish publishers for writing to stdout.
func (s *Stdout) Publish(data map[string]interface{}) {
	rawJSON, err := json.Marshal(data)
	if err != nil {
		s.log.Println("Stdout : Marshal ERROR :", err)
		return
	}

	var d map[string]interface{}
	if err := json.Unmarshal(rawJSON, &d); err != nil {
		s.log.Println("Stdout : Unmarshal ERROR :", err)
		return
	}

	// Add heap value into the data set.
	memStats, ok := (d["memstats"]).(map[string]interface{})
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
	s.log.Println("Stdout :\n", string(out))
}
