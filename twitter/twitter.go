package twitter

import (
	"errors"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
)

// Twitter login errors
var (
	ErrUnableToGetTwitterUser = errors.New("twitter: unable to get Twitter User")
)

// SuccessHandler is called when authentication via Twitter succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, user *twitter.User, token, tokenSecret string)
}

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, user *twitter.User, token, tokenSecret string)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, user *twitter.User, token, tokenSecret string) {
	f(w, req, user, token, tokenSecret)
}

// validateResponse returns an error if the given Twitter user, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *twitter.User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetTwitterUser
	}
	if user == nil || user.ID == 0 || user.IDStr == "" {
		return ErrUnableToGetTwitterUser
	}
	return nil
}
