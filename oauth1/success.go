package oauth1

import (
	"net/http"
)

// SuccessHandler is called when OAuth1 authentication succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, token, tokenSecret string)
}

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, token, tokenSecret string)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, token, tokenSecret string) {
	f(w, req, token, tokenSecret)
}
