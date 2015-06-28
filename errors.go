package login

import (
	"net/http"
)

// ErrorHandler handles login failues.
type ErrorHandler interface {
	ServeHTTP(w http.ResponseWriter, err error, code int)
}

// DefaultErrorHandler writes responses that pass-through the given error
// message and code.
var DefaultErrorHandler = &passthroughErrorHandler{}

type passthroughErrorHandler struct{}

func (e passthroughErrorHandler) ServeHTTP(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}
