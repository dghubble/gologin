package oauth1

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/internal"
	"github.com/dghubble/oauth1"
	"golang.org/x/net/context"
)

const (
	tempCookieName = "oauth1-temp-cookie"
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

// CookieTempHandler persists or retrieves the request token secret (temporary
// credentials). If the request token can be read from the ctx (login phase),
// the secret is set in a short-lived cookie to be read later. Otherwise
// (callback phase) the cookie is read to retrieve the request token secret
// and add it to the ctx.
// If the ctx contains no request token and the request has no temp cookie,
// the failure handler is called.
//
// Some OAuth1 providers (Twitter, Digits) do NOT require temp secrets to be
// kept between the login phase and callback phase. To implement those
// providers, use the EmptyTempHandler instead.
func CookieTempHandler(config gologin.CookieOptions, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		_, requestSecret, err := RequestTokenFromContext(ctx)
		if err == nil {
			// add request secret  to a short-lived cookie
			http.SetCookie(w, internal.NewCookie(tempCookieName, requestSecret, config))
			success.ServeHTTP(ctx, w, req)
			return
		}
		// read request secret from the short-lived cookie to add to ctx
		cookie, err := req.Cookie(tempCookieName)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithRequestToken(ctx, "", cookie.Value)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// EmptyTempHandler adds an empty request token secret to the ctx if none is
// present to support OAuth1 providers which do not require temp secrets to
// be kept between the login phase and callback phase.
func EmptyTempHandler(success ctxh.ContextHandler) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		_, _, err := RequestTokenFromContext(ctx)
		if err != nil {
			ctx = WithRequestToken(ctx, "", "")
		}
		success.ServeHTTP(ctx, w, req)
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
