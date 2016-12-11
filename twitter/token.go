package twitter

import (
	"fmt"
	"net/http"

	"github.com/dghubble/gologin"
	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/oauth1"
)

const (
	accessTokenField       = "twitterToken"
	accessTokenSecretField = "twitterTokenSecret"
)

// Errors for missing token or token secret form fields.
var (
	ErrMissingToken       = fmt.Errorf("twitter: missing token field %s", accessTokenField)
	ErrMissingTokenSecret = fmt.Errorf("twitter: missing token field %s", accessTokenSecretField)
)

// TokenHandler receives a Twitter access token/secret and calls Twitter
// verify_credentials to get the corresponding User. If successful, the access
// token/secret and User are added to the ctx and the success handler is
// called. Otherwise, the failure handler is called.
func TokenHandler(config *oauth1.Config, success, failure http.Handler) http.Handler {
	success = twitterHandler(config, success, failure)
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if req.Method != "POST" {
			ctx = gologin.WithError(ctx, fmt.Errorf("Method not allowed"))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		req.ParseForm()
		accessToken := req.PostForm.Get(accessTokenField)
		accessSecret := req.PostForm.Get(accessTokenSecretField)
		err := validateToken(accessToken, accessSecret)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = oauth1Login.WithAccessToken(ctx, accessToken, accessSecret)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateToken returns an error if the token or token secret is missing.
func validateToken(token, tokenSecret string) error {
	if token == "" {
		return ErrMissingToken
	}
	if tokenSecret == "" {
		return ErrMissingTokenSecret
	}
	return nil
}
