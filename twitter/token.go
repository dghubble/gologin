package twitter

import (
	"fmt"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/gologin"
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

// AuthClientSource is an interface for sources of oauth1 token authorized
// http.Client's. This interface avoids a hard dependency on a particular
// oauth1 implementation.
type AuthClientSource interface {
	GetClient(token, tokenSecret string) *http.Client
}

// TokenHandlerConfig configures a TokenHandler.
type TokenHandlerConfig struct {
	OAuth1Config AuthClientSource
	Success      SuccessHandler
	Failure      gologin.ErrorHandler
}

// TokenHandler receives a POSTed Twitter token/secret and verifies the Twitter
// credentials. If successful, handling is delegated to the SuccessHandler.
// Otherwise, the ErrorHandler is called.
type TokenHandler struct {
	oauth1Config AuthClientSource
	success      SuccessHandler
	failure      gologin.ErrorHandler
}

// NewTokenHandler returns a new TokenHandler.
func NewTokenHandler(config *TokenHandlerConfig) *TokenHandler {
	failure := config.Failure
	if failure == nil {
		failure = gologin.DefaultErrorHandler
	}
	return &TokenHandler{
		oauth1Config: config.OAuth1Config,
		success:      config.Success,
		failure:      failure,
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
	httpClient := h.oauth1Config.GetClient(accessToken, accessTokenSecret)
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
	h.success.ServeHTTP(w, req, user, accessToken, accessTokenSecret)
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
