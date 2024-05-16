package web

import "encoding/json"

type cors struct {
	Status string
}

// Encode implments the encoder interface.
func (c cors) Encode() ([]byte, string, error) {
	data, err := json.Marshal(c)
	return data, "application/json", err
}
