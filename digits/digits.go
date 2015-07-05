package digits

import (
	"errors"
	"net/http"

	"github.com/dghubble/go-digits/digits"
)

// Digits login errors.
var (
	ErrUnableToGetDigitsAccount = errors.New("digits: unable to get Digits account")
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

// validateResponse returns an error if the given Digits Account, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(account *digits.Account, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK || account == nil {
		return ErrUnableToGetDigitsAccount
	}
	if token := account.AccessToken; token.Token == "" || token.Secret == "" {
		// JSON deserialized Digits account is missing fields
		return ErrUnableToGetDigitsAccount
	}
	return nil
}
