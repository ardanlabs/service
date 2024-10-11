// Package rawapp provides an example of using a raw handler.
package rawapp

import (
	"encoding/json"
	"net/http"
)

func rawHandler(w http.ResponseWriter, r *http.Request) {
	status := struct {
		Status string
	}{
		Status: "RAW ENDPOINT",
	}

	json.NewEncoder(w).Encode(status)
}
