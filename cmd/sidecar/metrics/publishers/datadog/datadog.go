package datadog

import (
	"log"
	"sync"
	"time"
)

// Collector defines a contract a collector must support
// so a consumer can retrieve metrics.
type Collector interface {
	Collect() (map[string]interface{}, error)
}

// Console provides the ability to receive metrics
// from internal services using expvar.
type Console struct {
	collector Collector
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

// New creates a Console based consumer.
func New(collector Collector, interval time.Duration) (*Console, error) {
	con := Console{
		collector: collector,
		timer:     time.NewTimer(interval),
		shutdown:  make(chan struct{}),
	}

	con.wg.Add(1)
	go func() {
		defer con.wg.Done()
		for {
			con.timer.Reset(interval)
			select {
			case <-con.timer.C:
				con.publish()
			case <-con.shutdown:
				return
			}
		}
	}()

	return &con, nil
}

// Stop is used to shutdown the goroutine collecting metrics.
func (con *Console) Stop() {
	close(con.shutdown)
	con.wg.Wait()
}

// publish pulls the metrics and publishes them to the console.
func (con *Console) publish() {
	/*
		curl  -X POST -H "Content-type: application/json" \
		-d [JSON_DOC] \
		'https://app.datadoghq.com/api/v1/series?api_key=[KEY]'
	*/

	data, err := con.collector.Collect()
	if err != nil {
		log.Println(err)
		return
	}

	// out := marshal(data)

	log.Println(data)
}

// marshal handles the marshaling of the map to a datadog json document.
func marshal(data map[string]interface{}) string {
	/*
		{ \"series\" :
				[{\"metric\":\"test.metric\",
				\"points\":[[$currenttime, 20]],
				\"type\":\"gauge\",
				\"host\":\"test.example.com\",
				\"tags\":[\"environment:test\"]}
			]
		}
	*/

	type series struct {
		Metric string   `json:"metric"`
		Points []string `json:"points"`
		Type   string   `json:"type"`
		Host   string   `json:"host"`
		Tags   []string `json:"tags"`
	}
	type dog struct {
		series []series
	}

	return ""
}
