package gologin

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"golang.org/x/net/context"
)

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
