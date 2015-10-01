package tumblr

import (
	"errors"
	"net/http"
	"time"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/oauth1"
	"golang.org/x/net/context"
)

const (
	tempCookieName = "tumblr-temp-secret"
)

// Tumblr login errors
var (
	ErrUnableToGetTumblrUser = errors.New("tumblr: unable to get Tumblr User")
)

// LoginHandler handles Tumblr login requests by obtaining a request token,
// setting a temporary token secret cookie, and redirecting to the
// authorization URL.
func LoginHandler(config *oauth1.Config, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success := saveRequestSecret(oauth1Login.AuthRedirectHandler(config, failure), failure)
	return oauth1Login.LoginHandler(config, success, failure)
}

// CallbackHandler handles Tumblr callback requests by parsing the oauth token
// and verifier and adding the Tubmlr access token and User to the ctx. If
// authentication succeeds, handling delegates to the success handler,
// otherwise to the failure handler.
func CallbackHandler(config *oauth1.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success = verifyUser(config, success, failure)
	callback := oauth1Login.CallbackHandler(config, success, failure)
	return setRequestSecret(callback, failure)
}

// saveRequestSecret reads the request token secret from the ctx and sets it
// into a short lived cookie.
func saveRequestSecret(success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		_, requestSecret, err := oauth1Login.RequestTokenFromContext(ctx)
		http.SetCookie(w, newCookie(tempCookieName, requestSecret))
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// setRequestSecret parses the temporary cookie and adds the request token
// secret to the ctx.
func setRequestSecret(success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie(tempCookieName)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = oauth1Login.WithRequestToken(ctx, "", cookie.Value)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// TODO: cookie creation should be configurable
func newCookie(name, value string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   60,
		Secure:   false, //TODO
		HttpOnly: true,
	}
	d := time.Duration(60) * time.Second
	cookie.Expires = time.Now().Add(d)
	return cookie
}

func verifyUser(config *oauth1.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
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
		httpClient := config.Client(ctx, oauth1.NewToken(accessToken, accessSecret))
		tumblrClient := newClient(httpClient)
		user, resp, err := tumblrClient.UserInfo()
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
