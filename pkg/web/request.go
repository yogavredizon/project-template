package web

import (
	"io"
	"net/http"
)

type Decoder interface {
	Decode(data []byte) error
}

type validator interface {
	validate() error
}

func Decode(r *http.Request, d Decoder) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := d.Decode(data); err != nil {
		return err
	}

	if v, ok := d.(validator); ok {
		if err := v.validate(); err != nil {
			return err
		}
	}

	return nil
}
