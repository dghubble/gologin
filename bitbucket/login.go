package bitbucket

import (
	"errors"
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Bitbucket login errors
var (
	ErrUnableToGetBitbucketUser = errors.New("bitbucket: unable to get Bitbucket User")
)

// LoginHandler handles Bitbucket login requests by redirecting to the
// authorization URL.
func LoginHandler(config *oauth2.Config, stater oauth2Login.StateSource) ctxh.ContextHandler {
	return oauth2Login.LoginHandler(config, stater)
}

// CallbackHandler handles Bitbucket callback requests by parsing the auth
// code and state and adding the Bitbucket access token and User to the ctx.
// If authentication succeeds, handling delegates to the success handler,
// otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, stater oauth2Login.StateSource, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success = includeUser(config, success, failure)
	return oauth2Login.CallbackHandler(config, stater, success, failure)
}

// includeUser is a ContextHandler that gets the OAuth2 access token from the
// ctx to get the corresponding Bitbucket User. If successful, the User is
// added to the ctx and the success handler is called. Otherwise, the failure
// handler is called.
func includeUser(config *oauth2.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		accessToken, err := oauth2Login.AccessTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
		token := &oauth2.Token{AccessToken: accessToken}
		httpClient := config.Client(ctx, token)
		bitbucketClient := newClient(httpClient)
		user, resp, err := bitbucketClient.CurrentUser()
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

// validateResponse returns an error if the given Bitbucket User, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetBitbucketUser
	}
	if user == nil || user.Username == "" {
		return ErrUnableToGetBitbucketUser
	}
	return nil
}
