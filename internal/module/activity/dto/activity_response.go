package dto

import "encoding/json"

type Response struct {
	Status string `json:"status"`
}

func (r Response) Encode() ([]byte, string, error) {
	body, err := json.Marshal(r)
	return body, "application/json", err
}
