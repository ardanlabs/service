package checkapi

import "encoding/json"

type ready struct {
	Status string `json:"status"`
}

// Encode implments the encoder interface.
func (r ready) Encode() ([]byte, string, error) {
	b, err := json.Marshal(r)
	return b, "application/json", err
}
