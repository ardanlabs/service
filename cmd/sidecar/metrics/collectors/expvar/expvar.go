package expvar

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
			Timeout:   1 * time.Second,
		},
	}

	return &exp, nil
}

// Collect pulls the metrics from the configured host and returns JSON. This
// implements the Collector interface defined by publishers.
func (exp *Expvar) Collect() (map[string]interface{}, error) {
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
		msg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(msg))
	}

	data := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Add heap value into the data set.
	memStats, ok := (data["memstats"]).(map[string]interface{})
	if ok {
		data["heap"] = memStats["Alloc"]
	}

	// Remove uncessary keys.
	delete(data, "cmdline")
	delete(data, "memstats")

	return data, nil
}
