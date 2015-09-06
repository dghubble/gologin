// Package oauth2 provides handlers for OAuth2 login and callback requests.
package oauth2

import (
	"errors"
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Errors which may occur on login.
var (
	ErrInvalidState = errors.New("gologin: Invalid OAuth2 state parameter")
)

// LoginHandler handles OAuth2 login requests by redirecting to the
// authorization URL.
func LoginHandler(config *oauth2.Config, stater StateSource) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		authorizationURL := config.AuthCodeURL(stater.State())
		http.Redirect(w, req, authorizationURL, http.StatusFound)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// CallbackHandler handles OAuth2 callback requests by parsing the auth code
// and state and requesting an access token.
func CallbackHandler(config *oauth2.Config, stater StateSource, success ctxh.ContextHandler, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		authCode, state, err := parseCallback(req)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		if state != stater.State() {
			ctx = gologin.WithError(ctx, ErrInvalidState)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		// use the authorization code to get an access token
		token, err := config.Exchange(ctx, authCode)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithAccessToken(ctx, token.AccessToken)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// parseCallback parses the "code" and "state" parameters from the http.Request
// and returns them.
func parseCallback(req *http.Request) (authCode, state string, err error) {
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
