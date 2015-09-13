package gologin

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"golang.org/x/net/context"
)

// ErrorHandler handles login failues.
type ErrorHandler interface {
	ServeHTTP(w http.ResponseWriter, err error, code int)
}

// ErrorHandlerFunc is an adapter to allow an ordinary function to be used as
// an ErrorHandlerFunc.
type ErrorHandlerFunc func(w http.ResponseWriter, err error, code int)

func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, err error, code int) {
	f(w, err, code)
}

// DefaultErrorHandler writes responses that pass-through the given error
// message and code.
var DefaultErrorHandler = &passthroughErrorHandler{}

type passthroughErrorHandler struct{}

func (e passthroughErrorHandler) ServeHTTP(w http.ResponseWriter, err error, code int) {
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	http.Error(w, "", code)
}

// DefaultFailureHandler responds with a 400 status code and message parsed
// from the ctx.
var DefaultFailureHandler = ctxh.ContextHandlerFunc(failureHandler)

func failureHandler(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	err := ErrorFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Error(w, "", http.StatusBadRequest)
}
