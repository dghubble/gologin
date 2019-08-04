package tumblr

import (
	"errors"
	"net/http"

	"github.com/dghubble/gologin/v2"
	oauth1Login "github.com/dghubble/gologin/v2/oauth1"
	"github.com/dghubble/oauth1"
)

// Tumblr login errors
var (
	ErrUnableToGetTumblrUser = errors.New("tumblr: unable to get Tumblr User")
)

// LoginHandler handles Tumblr login requests by obtaining a request token,
// setting a temporary token secret cookie, and redirecting to the
// authorization URL.
func LoginHandler(config *oauth1.Config, cookieConfig gologin.CookieConfig, failure http.Handler) http.Handler {
	// oauth1.LoginHandler -> oauth1.CookieTempHander -> oauth1.AuthRedirectHandler
	success := oauth1Login.AuthRedirectHandler(config, failure)
	success = oauth1Login.CookieTempHandler(cookieConfig, success, failure)
	return oauth1Login.LoginHandler(config, success, failure)
}

// CallbackHandler handles Tumblr callback requests by parsing the oauth token
// and verifier and adding the Tubmlr access token and User to the ctx. If
// authentication succeeds, handling delegates to the success handler,
// otherwise to the failure handler.
func CallbackHandler(config *oauth1.Config, cookieConfig gologin.CookieConfig, success, failure http.Handler) http.Handler {
	// oauth1.CookieTempHandler -> oauth1.CallbackHandler -> TumblrHandler -> success
	success = tumblrHandler(config, success, failure)
	success = oauth1Login.CallbackHandler(config, success, failure)
	return oauth1Login.CookieTempHandler(cookieConfig, success, failure)
}

// tumblrHandler is a http.Handler that gets the OAuth1 access token from
// the ctx and obtains the Tumblr User. If successful, the User is added to
// the ctx and the success handler is called. Otherwise, the failure handler
// is called.
func tumblrHandler(config *oauth1.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		httpClient := config.Client(ctx, oauth1.NewToken(accessToken, accessSecret))
		tumblrClient := newClient(httpClient)
		user, resp, err := tumblrClient.UserInfo()
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

// validateResponse returns an error if the given Tumblr User, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetTumblrUser
	}
	if user == nil || user.Name == "" {
		return ErrUnableToGetTumblrUser
	}
	return nil
}
