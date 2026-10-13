package dto

import "encoding/json"

type Request struct {
	Content string `json:"content"`
}

func (r Request) Encode() ([]byte, string, error) {
	body, err := json.Marshal(r)
	return body, "application/json", err
}

func (r *Request) Decode(body []byte) error {
	if err := json.Unmarshal(body, r); err != nil {
		return err
	}

	return nil
}
