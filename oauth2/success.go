package oauth2

import (
	"net/http"
)

// SuccessHandler is called when OAuth2 authentication succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, accessToken string)
}

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, accessToken string)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, accessToken string) {
	f(w, req, accessToken)
}
