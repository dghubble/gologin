package google

import (
	"errors"
	"net/http"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"golang.org/x/oauth2"
	google "google.golang.org/api/oauth2/v2"
)

// Google login errors
var (
	ErrUnableToGetGoogleUser    = errors.New("google: unable to get Google User")
	ErrCannotValidateGoogleUser = errors.New("google: could not validate Google User")
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

// LoginHandler handles Google login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Google redirection URI requests and adds the Google
// access token and Userinfoplus to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = googleHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// googleHandler is a http.Handler that gets the OAuth2 Token from the ctx
// to get the corresponding Google Userinfoplus. If successful, the user info
// is added to the ctx and the success handler is called. Otherwise, the
// failure handler is called.
func googleHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
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
		googleService, err := google.New(httpClient)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		userInfoPlus, err := googleService.Userinfo.Get().Do()
		err = validateResponse(userInfoPlus, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithUser(ctx, userInfoPlus)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateResponse returns an error if the given Google Userinfoplus, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *google.Userinfoplus, err error) error {
	if err != nil {
		return ErrUnableToGetGoogleUser
	}
	if user == nil || user.Id == "" {
		return ErrCannotValidateGoogleUser
	}
	return nil
}
