// Package oauth2 provides a LoginHandler for OAuth2 login and callback
// requests.
package oauth2

import (
	"errors"
	"net/http"

	"github.com/dghubble/gologin"
	"golang.org/x/oauth2"
)

// Errors which may occur on login.
var (
	ErrInvalidState = errors.New("gologin: Invalid OAuth2 state parameter")
)

// LoginHandler handles OAuth2 login requests by redirecting to the
// authorization URL.
func LoginHandler(config *oauth2.Config, stater StateSource) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		authorizationURL := config.AuthCodeURL(stater.State())
		http.Redirect(w, req, authorizationURL, http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// CallbackHandler handles OAuth2 callback requests by reading the auth code
// and state and obtaining an access token.
func CallbackHandler(config *oauth2.Config, stater StateSource, success SuccessHandler, failure gologin.ErrorHandler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultErrorHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		authCode, state, err := validateCallback(req)
		if err != nil {
			failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		if state != stater.State() {
			failure.ServeHTTP(w, ErrInvalidState, http.StatusBadRequest)
			return
		}
		// use the authorization code to get an access token
		token, err := config.Exchange(oauth2.NoContext, authCode)
		if err != nil {
			failure.ServeHTTP(w, err, http.StatusBadRequest)
			return
		}
		success.ServeHTTP(w, req, token.AccessToken)
	}
	return http.HandlerFunc(fn)
}

func validateCallback(req *http.Request) (authCode, state string, err error) {
	// parse the raw query from the URL into req.Form
	err = req.ParseForm()
	if err != nil {
		return "", "", err
	}
	authCode = req.Form.Get("code")
	state = req.Form.Get("state")
	if authCode == "" || state == "" {
		return "", "", errors.New("callback did not receive a code or state")
	}
	return authCode, state, nil
}
