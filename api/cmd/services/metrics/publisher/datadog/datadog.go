// Package datadog provides support for publishing metrics to DD.
package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// Datadog provides the ability to publish metrics to Datadog.
type Datadog struct {
	log    *log.Logger
	apiKey string
	host   string
	tr     *http.Transport
	client http.Client
}

// New initializes Datadog access for publishing metrics.
func New(log *log.Logger, apiKey string, host string) *Datadog {
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

	d := Datadog{
		log:    log,
		apiKey: apiKey,
		host:   host,
		tr:     &tr,
		client: http.Client{
			Transport: &tr,
			Timeout:   1 * time.Second,
		},
	}

	return &d
}

// Publish handles the processing of metrics for deliver
// to the DataDog.
func (d *Datadog) Publish(data map[string]any) {
	doc, err := marshalDatadog(d.log, data)
	if err != nil {
		d.log.Println("datadog.publish :", err)
		return
	}

	if err := sendDatadog(d, doc); err != nil {
		d.log.Println("datadog.publish :", err)
		return
	}

	log.Println("datadog.publish : published :", string(doc))
}

// marshalDatadog converts the data map to datadog JSON document.
func marshalDatadog(log *log.Logger, data map[string]any) ([]byte, error) {
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
		Metric string   `json:"metric"`
		Points [][]any  `json:"points"`
		Type   string   `json:"type"`
		Host   string   `json:"host"`
		Tags   []string `json:"tags"`
	}

	// Populate the data into the data structure.
	var doc struct {
		Series []series `json:"series"`
	}
	for key, value := range data {
		switch value.(type) {
		case int, float64:
			doc.Series = append(doc.Series, series{
				Metric: env + "." + key,
				Points: [][]any{{"$currenttime", value}},
				Type:   mType,
				Host:   host,
				Tags:   []string{envTag},
			})
		}
	}

	// Convert the data into JSON.
	out, err := json.Marshal(doc)
	if err != nil {
		log.Println("datadog.publish : marshaling :", err)
		return nil, err
	}

	return out, nil
}

// sendDatadog sends data to the datadog servers.
func sendDatadog(d *Datadog, data []byte) error {
	url := fmt.Sprintf("%s?api_key=%s", d.host, d.apiKey)
	b := bytes.NewBuffer(data)

	r, err := http.NewRequest("POST", url, b)
	if err != nil {
		return err
	}

	resp, err := d.client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		out, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("status[%d] : %s", resp.StatusCode, out)
		}
		return fmt.Errorf("status[%d]", resp.StatusCode)
	}

	return nil
}
