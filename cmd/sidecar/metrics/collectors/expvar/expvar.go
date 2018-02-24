package expvar

import (
	"net"
	"net/http"
	"time"
)

// Expvar provides the ability to receive metrics
// from internal services using expvar.
type Expvar struct {
	host   string
	tr     *http.Transport
	client http.Client
}

// New creates a Expvar for collection metrics.
func New(host string) (*Expvar, error) {
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

	exp := Expvar{
		host: host,
		tr:   &tr,
		client: http.Client{
			Transport: &tr,
		},
	}

	return &exp, nil
}

// Collect pulls the metrics from the configured host. This implements
// the Collector interface defined by publishers.
func (exp *Expvar) Collect() string {
	return "metrics"
}
