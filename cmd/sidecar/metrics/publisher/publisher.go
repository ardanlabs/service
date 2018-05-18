package publisher

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Stdout publishers for writing to stdout
func Stdout(raw map[string]interface{}) {

	rawJSON, err := json.Marshal(raw)
	if err != nil {
		log.Println("Stdout : Marshal ERROR :", err)
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		log.Println("Stdout : Unmarshal ERROR :", err)
		return
	}

	// Add heap value into the data set.
	memStats, ok := (data["memstats"]).(map[string]interface{})
	if ok {
		data["heap"] = memStats["Alloc"]
	}

	// Remove uncessary keys.
	delete(data, "memstats")
	delete(data, "cmdline")

	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return
	}
	log.Println("Stdout :\n", string(out))
}

// =============================================================================

// Set of possible publisher types.
const (
	TypeConsole = "console"
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
	collector Collector
	publisher []Publisher
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

// New creates a Publish for consuming and publishing metrics.
func New(collector Collector, interval time.Duration, publisher ...Publisher) (*Publish, error) {
	p := Publish{
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
		log.Println(err)
		return
	}

	for _, pub := range p.publisher {
		pub(data)
	}
}
