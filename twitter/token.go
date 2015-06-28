package twitter

import (
	"errors"
	"net/http"

	"github.com/dghubble/go-login"
	"github.com/dghubble/go-twitter/twitter"
)

const (
	accessTokenField       = "twitterToken"
	accessTokenSecretField = "twitterTokenSecret"
)

// Errors indicating User credentials could not be verified.
var (
	ErrUnableToGetTwitterUser = errors.New("twitter: unable to get Twitter user")
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
	req.ParseForm()
	accessToken := req.PostForm.Get(accessTokenField)
	accessTokenSecret := req.PostForm.Get(accessTokenSecretField)
	httpClient := h.authConfig.GetClient(accessToken, accessTokenSecret)
	twitterClient := twitter.NewClient(httpClient)

	// verify Twitter access token
	accountVerifyParams := &twitter.AccountVerifyParams{
		IncludeEntities: twitter.Bool(false),
		SkipStatus:      twitter.Bool(true),
		IncludeEmail:    twitter.Bool(false),
	}
	user, resp, err := twitterClient.Accounts.VerifyCredentials(accountVerifyParams)
	err = validateAccountResponse(user, resp, err)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	h.success.ServeHTTP(w, req, user)
}

// SuccessHandler is called when authentication via Twitter succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, user *twitter.User)
}

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, user *twitter.User)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, user *twitter.User) {
	f(w, req, user)
}

// validateAccountResponse returns an error if the given Twitter user, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateAccountResponse(user *twitter.User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK || user == nil {
		return ErrUnableToGetTwitterUser
	}
	if user.ID == 0 || user.IDStr == "" {
		// JSON deserialized Twitter User is missing fields
		return ErrUnableToGetTwitterUser
	}
	return nil
}
