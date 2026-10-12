package errs

import (
	"encoding/json"
)

const (
	internalServerErrorMessage = "Terjadi kesalahan pada sistem"
)

type HttpError struct {
	Code    HttpCode `json:"code"`
	Message string   `json:"message"`
}

// HttpError.New setup new error to the response body with code as http status code.
// Internal server error will overwrite the message into InternalServerErrorMessage
func New(code HttpCode, msg string) *HttpError {
	return &HttpError{
		Message: msg,
		Code:    code,
	}
}

func (a *HttpError) Error() string {
	return a.Message
}

func (a HttpError) Encode() ([]byte, string, error) {
	if a.Code == InternalServerError {
		a.Message = internalServerErrorMessage
	}

	data, err := json.Marshal(a)
	return data, "application/json", err
}
