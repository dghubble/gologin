package digits

import (
	"fmt"
	"net/http"

	"github.com/dghubble/go-digits/digits"
	"github.com/dghubble/oauth1"
	"github.com/jbcjorge/gologin"
	oauth1Login "github.com/jbcjorge/gologin/oauth1"
)

const (
	accessTokenField       = "digitsToken"
	accessTokenSecretField = "digitsTokenSecret"
)

// Errors for missing token or token secret form fields.
var (
	ErrMissingToken       = fmt.Errorf("digits: missing Token form field %s", accessTokenField)
	ErrMissingTokenSecret = fmt.Errorf("digits: missing Token Secret form field %s", accessTokenSecretField)
)

// TokenHandler receives a Digits access token/secret and calls the Digits
// accounts endpoint to get the corresponding Account. If successful, the
// access token/secret and Account are added to the ctx and the success handler
// is called. Otherwise, the failure handler is called.
func TokenHandler(config *oauth1.Config, success, failure http.Handler) http.Handler {
	success = digitsHandler(config, success, failure)
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

// digitsHandler is a http.Handler that gets the OAuth1 access token from the
// ctx and calls the Digits accounts endpoint to get the corresponding Account.
// If successful, the Account is added to the ctx and the success handler is
// called. Otherwise, the failure handler is called.
func digitsHandler(config *oauth1.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		httpClient := config.Client(ctx, oauth1.NewToken(accessToken, accessSecret))
		digitsClient := digits.NewClient(httpClient)
		account, resp, err := digitsClient.Accounts.Account()
		err = validateResponse(account, resp, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithAccount(ctx, account)
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
