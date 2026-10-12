package errs

import "net/http"

type HttpCode int

const (
	ValidationError     HttpCode = http.StatusBadRequest
	InternalServerError HttpCode = http.StatusInternalServerError
)
