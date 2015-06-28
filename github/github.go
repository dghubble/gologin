package github

import (
	"errors"
	"net/http"

	"github.com/google/go-github/github"
)

// Github login errors
var (
	ErrUnableToGetGithubUser = errors.New("github: unable to get Github User")
)

// SuccessHandler is called when authentication via Github succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, user *github.User, accessToken string)
}

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, user *github.User, accessToken string)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, user *github.User, accessToken string) {
	f(w, req, user, accessToken)
}

// validateResponse returns an error if the given Github user, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *github.User, resp *github.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetGithubUser
	}
	if user == nil || user.ID == nil {
		return ErrUnableToGetGithubUser
	}
	return nil
}
