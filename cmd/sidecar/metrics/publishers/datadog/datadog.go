package datadog

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Collector defines a contract a collector must support
// so a consumer can retrieve metrics.
type Collector interface {
	Collect() (map[string]interface{}, error)
}

// Datadog provides the ability to receive metrics
// from internal services using expvar.
type Datadog struct {
	collector Collector
	wg        sync.WaitGroup
	timer     *time.Timer
	shutdown  chan struct{}
}

// New creates a Datadog based consumer.
func New(collector Collector, interval time.Duration) (*Datadog, error) {
	dg := Datadog{
		collector: collector,
		timer:     time.NewTimer(interval),
		shutdown:  make(chan struct{}),
	}

	dg.wg.Add(1)
	go func() {
		defer dg.wg.Done()
		for {
			dg.timer.Reset(interval)
			select {
			case <-dg.timer.C:
				dg.publish()
			case <-dg.shutdown:
				return
			}
		}
	}()

	return &dg, nil
}

// Stop is used to shutdown the goroutine collecting metrics.
func (dg *Datadog) Stop() {
	close(dg.shutdown)
	dg.wg.Wait()
}

// publish pulls the metrics and publishes them to the Datadog.
func (dg *Datadog) publish() {
	/*
		curl  -X POST -H "Content-type: application/json" \
		-d [JSON_DOC] \
		'https://app.datadoghq.com/api/v1/series?api_key=[KEY]'
	*/

	data, err := dg.collector.Collect()
	if err != nil {
		log.Println(err)
		return
	}

	out, err := marshal(data)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(out)
}

// marshal handles the marshaling of the map to a datadog json document.
func marshal(data map[string]interface{}) (string, error) {
	/*
		{ "series" : [
				{
					"metric":"test.metric",
					"points": [
						[
							$currenttime,
							20
						]
					],
					"type":"gauge",
					"host":"test.example.com",
					"tags": [
						"environment:test"
					]
				}
			]
		}
	*/

	// Extract the base keys/values.
	mType := "gauge"
	host, ok := data["host"].(string)
	if !ok {
		host = "unknown"
	}
	env := "dev"
	if host != "localhost" {
		env = "prod"
	}
	envTag := "environment:" + env

	// Define the Datadog data format.
	type series struct {
		Metric string          `json:"metric"`
		Points [][]interface{} `json:"points"`
		Type   string          `json:"type"`
		Host   string          `json:"host"`
		Tags   []string        `json:"tags"`
	}
	type dog struct {
		Series []series `json:"series"`
	}

	// Populate the data into the data structure.
	var d dog
	for key, value := range data {
		switch value.(type) {
		case int, float64:
			d.Series = append(d.Series, series{
				Metric: env + "." + key,
				Points: [][]interface{}{[]interface{}{"$currenttime", value}},
				Type:   mType,
				Host:   host,
				Tags:   []string{envTag},
			})
		}
	}

	// Convert the data into JSON.
	out, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		return "", err
	}

	return string(out), nil
}
