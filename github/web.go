package github

import (
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// LoginHandler handles Github login requests by redirecting to the
// authorization URL.
func LoginHandler(config *oauth2.Config, stater oauth2Login.StateSource) ctxh.ContextHandler {
	return oauth2Login.LoginHandler(config, stater)
}

// CallbackHandler handles Github callback requests by parsing the auth code
// and state and adding the Github access token and User to the ctx. If
// authentication succeeds, handling delegates to the success handler,
// otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, stater oauth2Login.StateSource, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success = IncludeUser(config, success, failure)
	return oauth2Login.CallbackHandler(config, stater, success, failure)
}

// IncludeUser is a ContextHandler that gets the OAuth2 access token from the
// ctx to get the corresponding Github User. If successful, the User is added
// to the ctx and the success handler is called. Otherwise the failure handler
// is called.
func IncludeUser(config *oauth2.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
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
		githubClient := github.NewClient(httpClient)
		user, resp, err := githubClient.Users.Get("")
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
