package web

import "encoding/json"

type cors struct {
	Status string
}

func (c cors) Encode() ([]byte, string, error) {
	b, err := json.Marshal(c)
	return b, "application/json", err
}
