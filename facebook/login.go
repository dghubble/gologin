package facebook

import (
	"errors"
	"net/http"

	"github.com/jbcjorge/gologin"
	oauth2Login "github.com/jbcjorge/gologin/oauth2"
	"golang.org/x/oauth2"
)

// Facebook login errors
var (
	ErrUnableToGetFacebookUser = errors.New("facebook: unable to get Facebook User")
)

// StateHandler checks for a state cookie. If found, the state value is read
// and added to the ctx. Otherwise, a non-guessable value is added to the ctx
// and to a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a http.Handler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func StateHandler(config gologin.CookieConfig, success http.Handler) http.Handler {
	return oauth2Login.StateHandler(config, success)
}

// LoginHandler handles Facebook login requests by reading the state value
// from the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Facebook redirection URI requests and adds the
// Facebook access token and User to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure
// handler.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = facebookHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// facebookHandler is a http.Handler that gets the OAuth2 Token from the ctx
// to get the corresponding Facebook User. If successful, the user is added to
// the ctx and the success handler is called. Otherwise, the failure handler
// is called.
func facebookHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := oauth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		httpClient := config.Client(ctx, token)
		facebookService := newClient(httpClient)
		user, resp, err := facebookService.Me()
		err = validateResponse(user, resp, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithUser(ctx, user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateResponse returns an error if the given Facebook User, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetFacebookUser
	}
	if user == nil || user.ID == "" {
		return ErrUnableToGetFacebookUser
	}
	return nil
}
