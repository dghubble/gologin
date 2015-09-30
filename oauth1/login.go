// Package oauth1 provides handles for OAuth1 login and callback requests.
package oauth1

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	"github.com/dghubble/oauth1"
	"golang.org/x/net/context"
)

// LoginHandler handles OAuth1 login requests by obtaining a request token and
// secret (temporary credentials) and adding it to the ctx. If successful,
// handling delegates to the success handler, otherwise to the failure handler.
//
// Typically, the success handler is an AuthRedirectHandler or a handler which
// stores the request token secret.
func LoginHandler(config *oauth1.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		requestToken, requestSecret, err := config.RequestToken()
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithRequestToken(ctx, requestToken, requestSecret)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// AuthRedirectHandler reads the request token from the ctx and redirects
// to the authorization URL.
func AuthRedirectHandler(config *oauth1.Config, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		requestToken, _, err := RequestTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
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
// and verifier, reading the request token secret from the ctx, then obtaining
// an access token and adding it to the ctx.
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

		// upstream handler should add the request token secret from the login step
		_, requestSecret, err := RequestTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}

		accessToken, accessSecret, err := config.AccessToken(requestToken, requestSecret, verifier)
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
