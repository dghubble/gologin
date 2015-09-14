package digits

import (
	"net/http"

	"github.com/dghubble/go-digits/digits"
)

// SuccessHandler is called when authentication via Digits succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, account *digits.Account)
}

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, account *digits.Account)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, account *digits.Account) {
	f(w, req, account)
}
