package web

import "encoding/json"

type cors struct {
	Status string
}

func (c cors) Encode() ([]byte, error) {
	return json.Marshal(c)
}
