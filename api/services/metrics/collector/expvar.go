// Package collector is a simple collector for
package collector

import (
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-json-experiment/json"
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
			Timeout:   1 * time.Second,
		},
	}

	return &exp, nil
}

// Collect captures metrics on the host configure to this endpoint.
func (exp *Expvar) Collect() (map[string]any, error) {
	req, err := http.NewRequest("GET", exp.host, nil)
	if err != nil {
		return nil, err
	}

	resp, err := exp.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(msg))
	}

	data := make(map[string]any)
	if err := json.UnmarshalRead(resp.Body, &data); err != nil {
		return nil, err
	}

	return data, nil
}
