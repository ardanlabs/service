package trace

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

// Exporter provides support to batch spans and send them
// to the sidecar for processing.
type Exporter struct {
	host      string
	batchSize int
	interval  time.Duration
	tr        *http.Transport
	client    http.Client
	batch     []*trace.SpanData
	mu        sync.Mutex
	timer     *time.Timer
}

// NewExporter creates an exporter for use.
func NewExporter(host string, batchSize int, interval time.Duration) *Exporter {
	tr := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	e := Exporter{
		host:      host,
		batchSize: batchSize,
		interval:  interval,
		tr:        &tr,
		client: http.Client{
			Transport: &tr,
		},
		batch: make([]*trace.SpanData, 0, batchSize),
		timer: time.NewTimer(interval),
	}

	return &e
}

// ExportSpan is called by goroutines when saving spans via
// the opentracing API.
func (e *Exporter) ExportSpan(span *trace.SpanData) {
	sendBatch := e.save(span)
	if sendBatch != nil {
		go e.send(sendBatch)
	}
}

// Saves the span data to the batch. If the batch should be sent,
// returns a batch to send.
func (e *Exporter) save(span *trace.SpanData) []*trace.SpanData {
	var sendBatch []*trace.SpanData

	e.mu.Lock()
	{
		// We want to append this new span to the collection.
		e.batch = append(e.batch, span)

		// Do we need to send the current batch?
		switch {
		case len(e.batch) == e.batchSize:

			// We hit the batch size. Now save the current
			// batch for sending and start a new batch.
			sendBatch = e.batch
			e.batch = make([]*trace.SpanData, 0, e.batchSize)
			e.timer.Reset(e.interval)

		default:

			// We did not hit the batch size but maybe send what
			// we have based on time.
			select {
			case <-e.timer.C:

				// The time has expired so save the current
				// batch for sending and start a new batch.
				sendBatch = e.batch
				e.batch = make([]*trace.SpanData, 0, e.batchSize)

			// It's not time yet, just move on.
			default:
			}
		}
	}
	e.mu.Unlock()

	return sendBatch
}

// send uses HTTP to send the data to the tracing sidecare for processing.
func (e *Exporter) send(sendBatch []*trace.SpanData) {
	log.Println("******************** SENDING **********************")
	data, err := json.MarshalIndent(sendBatch, "", "  ")
	if err != nil {
		log.Println(err)
	} else {
		log.Println(string(data))
	}
	log.Println("******************** SENDING **********************")
}
