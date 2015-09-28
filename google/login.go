package google

import (
	"errors"
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	google "github.com/google/google-api-go-client/oauth2/v2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Google login errors
var (
	ErrUnableToGetGoogleUser    = errors.New("google: unable to get Google User")
	ErrCannotValidateGoogleUser = errors.New("google: could not validate Google User")
)

// LoginHandler handles Google login requests by redirecting to the
// authorization URL.
func LoginHandler(config *oauth2.Config, stater oauth2Login.StateSource) ctxh.ContextHandler {
	return oauth2Login.LoginHandler(config, stater)
}

// CallbackHandler handles Google callback requests by parsing the auth code
// and state and adding the Google access token and User to the ctx. If
// authentication succeeds, handling delegates to the success handler,
// otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, stater oauth2Login.StateSource, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success = includeUser(config, success, failure)
	return oauth2Login.CallbackHandler(config, stater, success, failure)
}

// includeUser is a ContextHandler that gets the OAuth2 access token from the
// ctx to get the corresponding Google Userinfoplus. If successful, the user
// info is added to the ctx and the success handler is called. Otherwise, the
// failure handler is called.
func includeUser(config *oauth2.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		accessToken, err := oauth2Login.AccessTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		token := &oauth2.Token{AccessToken: accessToken}
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
