package traexp

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"go.opencensus.io/trace"
)

// Exporter is a trace exporter that posts the exported data
// to the specified HTTP based host.
type Exporter struct {
	host   string
	tr     *http.Transport
	client http.Client
}

// New returns an Exporter for use to communicate with the tracing sidecar.
func New(host string) *Exporter {
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
		host: host,
		tr:   &tr,
		client: http.Client{
			Transport: &tr,
		},
	}

	return &e
}

// ExportSpan posts the traces to the configured host.
func (e *Exporter) ExportSpan(vd *trace.SpanData) {
	log.Println("trace : Export Spans : Started")

	d, err := json.MarshalIndent(vd, "", "  ")
	if err != nil {
		log.Println("trace : ERROR :", err)
		return
	}

	req, err := http.NewRequest("POST", e.host, bytes.NewBuffer(d))
	if err != nil {
		log.Println("trace : ERROR :", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		resp, err := e.client.Do(req)
		if err != nil {
			ch <- err
			return
		}

		defer resp.Body.Close()
		ch <- nil
	}()

	select {
	case <-ctx.Done():
		e.tr.CancelRequest(req)

	case err := <-ch:
		if err != nil {
			log.Println("trace : ERROR :", err)
		}
	}

	log.Println("trace : Export Spans : Completed")
}
