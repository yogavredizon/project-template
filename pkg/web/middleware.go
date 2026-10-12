package web

import "slices"

type MidFun func(HandlerFunc) HandlerFunc

func wrapMiddleware(midds []MidFun, handler HandlerFunc) HandlerFunc {
	for _, m := range slices.Backward(midds) {
		if m != nil {
			handler = m(handler)
		}
	}

	return handler
}
