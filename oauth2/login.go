package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	stateCookie = "state-cookie"
)

// Errors which may occur on login.
var (
	ErrInvalidState = errors.New("oauth2: Invalid OAuth2 state parameter")
)

// StateHandler checks for a temporary state cookie. If found, the state value
// is read from it and added to the ctx. Otherwise, a temporary state cookie
// is written and the corresponding state value is added to the ctx.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection.
func StateHandler(success ctxh.ContextHandler) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie(stateCookie)
		if err == nil {
			ctx = WithState(ctx, cookie.Value)
		} else {
			val := randomState()
			http.SetCookie(w, newCookie(stateCookie, val))
			ctx = WithState(ctx, val)
		}
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// LoginHandler handles OAuth2 login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		state, err := StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		authURL := config.AuthCodeURL(state)
		http.Redirect(w, req, authURL, http.StatusFound)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// CallbackHandler handles OAuth2 redirection URI requests by parsing the auth
// code and state, comparing with the state value from the ctx, and obtaining
// an OAuth2 Token.
func CallbackHandler(config *oauth2.Config, success ctxh.ContextHandler, failure ctxh.ContextHandler) ctxh.ContextHandler {
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
		ownerState, err := StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		if state != ownerState || state == "" {
			ctx = gologin.WithError(ctx, ErrInvalidState)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		// use the authorization code to get a Token
		token, err := config.Exchange(ctx, authCode)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithAccessToken(ctx, token)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
}

// Returns a base64 encoded random 32 byte string.
func randomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
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
		return "", "", errors.New("oauth2: Request missing code or state")
	}
	return authCode, state, nil
}
