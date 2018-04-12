package trace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

// Error variables for factory validation.
var (
	ErrLoggerNotProvided = errors.New("logger not provided")
	ErrHostNotProvided   = errors.New("host not provided")
)

// Log provides support for logging inside this package.
// Unfortunately, the opentrace API calls into the ExportSpan
// function directly with no means to pass user defined arguments.
type Log func(format string, v ...interface{})

// Exporter provides support to batch spans and send them
// to the sidecar for processing.
type Exporter struct {
	log          Log               // Handler function for logging.
	host         string            // IP:port of the sidecare consuming the trace data.
	batchSize    int               // Size of the batch of spans before sending.
	sendInterval time.Duration     // Time to send a batch if batch size is not met.
	sendTimeout  time.Duration     // Time to wait for the sidecar to respond on send.
	client       http.Client       // Provides APIs for performing the http send.
	batch        []*trace.SpanData // Maintains the batch of span data to be sent.
	mu           sync.Mutex        // Provide synchronization to access the batch safely.
	timer        *time.Timer       // Signals when the sendInterval is met.
}

// NewExporter creates an exporter for use.
func NewExporter(log Log, host string, batchSize int, sendInterval, sendTimeout time.Duration) (*Exporter, error) {
	if log == nil {
		return nil, ErrLoggerNotProvided
	}
	if host == "" {
		return nil, ErrHostNotProvided
	}

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
		log:          log,
		host:         host,
		batchSize:    batchSize,
		sendInterval: sendInterval,
		sendTimeout:  sendTimeout,
		client: http.Client{
			Transport: &tr,
		},
		batch: make([]*trace.SpanData, 0, batchSize),
		timer: time.NewTimer(sendInterval),
	}

	return &e, nil
}

// Close sends the remaining spans that have not been sent yet.
func (e *Exporter) Close() (int, error) {
	var sendBatch []*trace.SpanData
	e.mu.Lock()
	{
		sendBatch = e.batch
	}
	e.mu.Unlock()

	err := e.send(sendBatch)
	if err != nil {
		return len(sendBatch), err
	}

	return len(sendBatch), nil
}

// ExportSpan is called by goroutines when saving spans via
// the opentracing API.
func (e *Exporter) ExportSpan(span *trace.SpanData) {
	sendBatch := e.save(span)
	if sendBatch != nil {
		go func() {
			e.log("trace : Exporter : ExportSpan : Sending Batch[%d]", len(sendBatch))
			if err := e.send(sendBatch); err != nil {
				e.log("trace : Exporter : ExportSpan : ERROR : %v", err)
			}
		}()
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
			e.timer.Reset(e.sendInterval)

		default:

			// We did not hit the batch size but maybe send what
			// we have based on time.
			select {
			case <-e.timer.C:

				// The time has expired so save the current
				// batch for sending and start a new batch.
				sendBatch = e.batch
				e.batch = make([]*trace.SpanData, 0, e.batchSize)
				e.timer.Reset(e.sendInterval)

			// It's not time yet, just move on.
			default:
			}
		}
	}
	e.mu.Unlock()

	return sendBatch
}

// send uses HTTP to send the data to the tracing sidecare for processing.
func (e *Exporter) send(sendBatch []*trace.SpanData) error {
	data, err := json.Marshal(sendBatch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", e.host, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(req.Context(), e.sendTimeout)
	defer cancel()
	req = req.WithContext(ctx)

	ch := make(chan error)
	go func() {
		resp, err := e.client.Do(req)
		if err != nil {
			ch <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- fmt.Errorf("error on call : status[%s]", resp.Status)
				return
			}
			ch <- fmt.Errorf("error on call : status[%s] : %s", resp.Status, string(data))
			return
		}

		ch <- nil
	}()

	return <-ch
}
