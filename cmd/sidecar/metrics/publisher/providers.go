package publisher

import (
	"encoding/json"
	"log"
)

// Set of possible publisher types.
const (
	TypeConsole = "console"
	TypeDatadog = "datadog"
)

// Console handles the processing of metrics for deliver
// to the console.
func Console(data map[string]interface{}) {
	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return
	}
	log.Println(string(out))
}

// Datadog handles the processing of metrics for deliver
// to the DataDog.
func Datadog(data map[string]interface{}) {
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
		return
	}
	log.Println(string(out))
}
