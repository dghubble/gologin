package twitter

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/gologin"
	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/oauth1"
	"golang.org/x/net/context"
)

// LoginHandler handles Twitter login requests by obtaining a request token and
// redirecting to the authorization URL.
func LoginHandler(config *oauth1.Config, failure ctxh.ContextHandler) ctxh.ContextHandler {
	return oauth1Login.LoginHandler(config, failure)
}

// CallbackHandler handles Twitter callback requests by parsing the oauth token
// and verifier and adding the Twitter access token and User to the ctx. If
// authentication succeeds, handling delegates to the success handler,
// otherwise to the failure handler.
func CallbackHandler(config *oauth1.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success = VerifyUser(config, success, failure)
	return oauth1Login.CallbackHandler(config, success, failure)
}

// VerifyUser is a ContextHandler that gets the OAuth1 Access Token from the
// ctx and calls Twitter verify_credentials to get the corresponding User.
// If successful, the User is added to the ctx and the success handler is
// called. Otherwise the failure handler is called.
func VerifyUser(config *oauth1.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		httpClient := config.Client(oauth1.NewToken(accessToken, accessSecret))
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
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithUser(ctx, user)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}
