package authapi

import "encoding/json"

type token struct {
	Token string `json:"token"`
}

// Encode implments the encoder interface.
func (t token) Encode() ([]byte, string, error) {
	b, err := json.Marshal(t)
	return b, "application/json", err
}
