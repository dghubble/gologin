// Package oauth1 provides handles for OAuth1 login and callback requests.
package oauth1

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	"github.com/dghubble/oauth1"
	"golang.org/x/net/context"
)

// TODO: generalize oauth1 to interface
// TODO: remove some Twitter specific aspects

// LoginHandler handles OAuth1 login requests by obtaining a request token and
// redirecting to the authorization URL.
func LoginHandler(config *oauth1.Config, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		requestToken, _, err := config.RequestToken()
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		// Twitter does not require the oauth token secret be saved
		authorizationURL, err := config.AuthorizationURL(requestToken)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		http.Redirect(w, req, authorizationURL.String(), http.StatusFound)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// CallbackHandler handles OAuth1 callback requests by parsing the oauth token
// and verifier, then obtaining an access token.
func CallbackHandler(config *oauth1.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		requestToken, verifier, err := oauth1.ParseAuthorizationCallback(req)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		// Twitter AccessToken endpoint does not require the auth header to be signed
		// No need to lookup the (temporary) RequestToken secret token.
		accessToken, accessSecret, err := config.AccessToken(requestToken, "", verifier)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithAccessToken(ctx, accessToken, accessSecret)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}
