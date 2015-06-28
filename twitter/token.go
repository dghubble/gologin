package twitter

import (
	"fmt"
	"net/http"

	login "github.com/dghubble/go-login"
	"github.com/dghubble/go-twitter/twitter"
)

const (
	accessTokenField       = "twitterToken"
	accessTokenSecretField = "twitterTokenSecret"
)

// TokenHandler errors.
var (
	ErrMissingToken       = fmt.Errorf("twitter: missing token field %s", accessTokenField)
	ErrMissingTokenSecret = fmt.Errorf("twitter: missing token field %s", accessTokenSecretField)
)

// AuthClientSource is an interface for sources of oauth1 token authorized
// http.Client's. This interface avoids a hard dependency on a particular
// oauth1 implementation.
type AuthClientSource interface {
	GetClient(token, tokenSecret string) *http.Client
}

// TokenHandlerConfig configures a TokenHandler.
type TokenHandlerConfig struct {
	AuthConfig AuthClientSource
	Success    SuccessHandler
	Failure    login.ErrorHandler
}

// TokenHandler receives a POSTed Twitter token/secret and verifies the Twitter
// credentials. If successful, handling is delegated to the SuccessHandler.
// Otherwise, the ErrorHandler is called.
type TokenHandler struct {
	authConfig AuthClientSource
	success    SuccessHandler
	failure    login.ErrorHandler
}

// NewTokenHandler returns a new TokenHandler.
func NewTokenHandler(config *TokenHandlerConfig) *TokenHandler {
	return &TokenHandler{
		authConfig: config.AuthConfig,
		success:    config.Success,
		failure:    config.Failure,
	}
}

// ServeHTTP receives a POSTed Twitter token/secret and verifies the Twitter
// credentials. If successful, handling is delegated to the SuccessHandler.
// Otherwise, the ErrorHandler is called.
func (h *TokenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		h.failure.ServeHTTP(w, nil, http.StatusMethodNotAllowed)
		return
	}
	req.ParseForm()
	accessToken := req.PostForm.Get(accessTokenField)
	accessTokenSecret := req.PostForm.Get(accessTokenSecretField)
	err := validateToken(accessToken, accessTokenSecret)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	// verify Twitter access token
	httpClient := h.authConfig.GetClient(accessToken, accessTokenSecret)
	twitterClient := twitter.NewClient(httpClient)
	accountVerifyParams := &twitter.AccountVerifyParams{
		IncludeEntities: twitter.Bool(false),
		SkipStatus:      twitter.Bool(true),
		IncludeEmail:    twitter.Bool(false),
	}
	user, resp, err := twitterClient.Accounts.VerifyCredentials(accountVerifyParams)
	err = validateResponse(user, resp, err)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	h.success.ServeHTTP(w, req, user)
}

// validateToken returns an error if the token or token secret is missing.
func validateToken(token, secret string) error {
	if token == "" {
		return ErrMissingToken
	}
	if secret == "" {
		return ErrMissingTokenSecret
	}
	return nil
}
