package publisher

import (
	"encoding/json"
	"log"
)

// Console handles the processing of metrics for deliver
// to the console.
func Console(data map[string]interface{}) {
	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return
	}
	log.Println("console :\n", string(out))
}
