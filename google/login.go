package google

import (
	"errors"
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	google "google.golang.org/api/oauth2/v2"
)

// Google login errors
var (
	ErrUnableToGetGoogleUser    = errors.New("google: unable to get Google User")
	ErrCannotValidateGoogleUser = errors.New("google: could not validate Google User")
)

// StateHandler checks for a temporary state cookie. If found, the state value
// is read from it and added to the ctx. Otherwise, a temporary state cookie
// is written and the corresponding state value is added to the ctx.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection.
func StateHandler(success ctxh.ContextHandler) ctxh.ContextHandler {
	return oauth2Login.StateHandler(success)
}

// LoginHandler handles Google login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure ctxh.ContextHandler) ctxh.ContextHandler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Google redirection URI requests and adds the Google
// access token and Userinfoplus to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success = googleHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// googleHandler is a ContextHandler that gets the OAuth2 Token from the ctx
// to get the corresponding Google Userinfoplus. If successful, the user info
// is added to the ctx and the success handler is called. Otherwise, the
// failure handler is called.
func googleHandler(config *oauth2.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		token, err := oauth2Login.AccessTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		httpClient := config.Client(ctx, token)
		googleService, err := google.New(httpClient)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		userInfoPlus, err := googleService.Userinfo.Get().Do()
		err = validateResponse(userInfoPlus, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		ctx = WithUser(ctx, userInfoPlus)
		success.ServeHTTP(ctx, w, req)
	}
	return ctxh.ContextHandlerFunc(fn)
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
