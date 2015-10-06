package github

import (
	"errors"
	"net/http"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// Github login errors
var (
	ErrUnableToGetGithubUser = errors.New("github: unable to get Github User")
)

// StateHandler checks for a state cookie. If found, the state value is read
// and added to the ctx. Otherwise, a non-guessable value is added to the ctx
// and to a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a ContextHandler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func StateHandler(success ctxh.ContextHandler, opts ...gologin.CookieOptions) ctxh.ContextHandler {
	return oauth2Login.StateHandler(success, opts...)
}

// LoginHandler handles Github login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure ctxh.ContextHandler) ctxh.ContextHandler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Github redirection URI requests and adds the Github
// access token and User to the ctx. If authentication succeeds, handling
// delegates to the success handler, otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	success = githubHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// githubHandler is a ContextHandler that gets the OAuth2 Token from the ctx to
// get the corresponding Github User. If successful, the User is added to the
// ctx and the success handler is called. Otherwise, the failure handler is
// called.
func githubHandler(config *oauth2.Config, success, failure ctxh.ContextHandler) ctxh.ContextHandler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		token, err := oauth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(ctx, w, req)
			return
		}
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

// validateResponse returns an error if the given Github user, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *github.User, resp *github.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetGithubUser
	}
	if user == nil || user.ID == nil {
		return ErrUnableToGetGithubUser
	}
	return nil
}
