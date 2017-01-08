package twitter

import (
	"errors"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/gologin"
	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/oauth1"
)

// Twitter login errors
var (
	ErrUnableToGetTwitterUser = errors.New("twitter: unable to get Twitter User")
)

// LoginHandler handles Twitter login requests by obtaining a request token and
// redirecting to the authorization URL.
func LoginHandler(config *oauth1.Config, failure http.Handler) http.Handler {
	// oauth1.LoginHandler -> oauth1.AuthRedirectHandler
	success := oauth1Login.AuthRedirectHandler(config, failure)
	return oauth1Login.LoginHandler(config, success, failure)
}

// CallbackHandler handles Twitter callback requests by parsing the oauth token
// and verifier and adding the Twitter access token and User to the ctx. If
// authentication succeeds, handling delegates to the success handler,
// otherwise to the failure handler.
func CallbackHandler(config *oauth1.Config, success, failure http.Handler) http.Handler {
	// oauth1.EmptyTempHandler -> oauth1.CallbackHandler -> TwitterHandler -> success
	success = twitterHandler(config, success, failure)
	success = oauth1Login.CallbackHandler(config, success, failure)
	return oauth1Login.EmptyTempHandler(success)
}

// twitterHandler is a http.Handler that gets the OAuth1 access token from
// the ctx and calls Twitter verify_credentials to get the corresponding User.
// If successful, the User is added to the ctx and the success handler is
// called. Otherwise, the failure handler is called.
func twitterHandler(config *oauth1.Config, success, failure http.Handler) http.Handler {
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
		twitterClient := twitter.NewClient(httpClient)
		accountVerifyParams := &twitter.AccountVerifyParams{
			IncludeEntities: twitter.Bool(false),
			SkipStatus:      twitter.Bool(true),
			IncludeEmail:    twitter.Bool(false),
		}
		user, resp, err := twitterClient.Accounts.VerifyCredentials(accountVerifyParams)
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
