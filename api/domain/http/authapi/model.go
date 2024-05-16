package authapi

import "encoding/json"

type token struct {
	Token string `json:"token"`
}

// Encode implments the encoder interface.
func (t token) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	return data, "application/json", err
}
