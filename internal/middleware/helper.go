package middleware

import (
	"github.com/yogavredizon/log-analyzer/pkg/web"
)

func checkIsError(e web.Encoder) error {
	err, ok := e.(error)
	if ok {
		return err
	}

	return nil
}
