// Package login handles Digits token based logins, typically for mobile clients.
package login

import (
	"fmt"
	"net/http"

	"github.com/dghubble/go-digits/digits"
)

const (
	accessTokenKey       = "digitsToken"
	accessTokenSecretKey = "digitsTokenSecret"
)

// Errors for missing token or token secret form fields.
var (
	ErrMissingToken       = fmt.Errorf("digits: missing Token form field %s", accessTokenKey)
	ErrMissingTokenSecret = fmt.Errorf("digits: missing Token Secret form field %s", accessTokenSecretKey)
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
	Failure    ErrorHandler
}

// TokenHandler receives a POSTed Digits token/secret and fetches the Digits
// Account. If successful, handling is delegated to the SuccessHandler.
// Otherwise, the ErrorHandler is called.
type TokenHandler struct {
	authConfig AuthClientSource
	success    SuccessHandler
	failure    ErrorHandler
}

// NewTokenHandler returns a new TokenHandler.
func NewTokenHandler(config *TokenHandlerConfig) *TokenHandler {
	return &TokenHandler{
		authConfig: config.AuthConfig,
		success:    config.Success,
		failure:    config.Failure,
	}
}

// ServeHTTP receives a POSTed Digits token/secret and fetches the Digits
// Account. If successful, handling is delegated to the SuccessHandler.
// Otherwise, the ErrorHandler is called.
func (h *TokenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		h.failure.ServeHTTP(w, nil, http.StatusMethodNotAllowed)
		return
	}
	req.ParseForm()
	accessToken := req.PostForm.Get(accessTokenKey)
	accessTokenSecret := req.PostForm.Get(accessTokenSecretKey)
	err := validateToken(accessToken, accessTokenSecret)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	httpClient := h.authConfig.GetClient(accessToken, accessTokenSecret)
	digitsClient := digits.NewClient(httpClient)

	// fetch Digits Account
	account, resp, err := digitsClient.Accounts.Account()
	err = validateAccountResponse(account, resp, err)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	h.success.ServeHTTP(w, req, account)
}

// validateToken returns an error if the token or token secret is missing.
func validateToken(accessToken, accessTokenSecret string) error {
	if accessToken == "" {
		return ErrMissingToken
	}
	if accessTokenSecret == "" {
		return ErrMissingTokenSecret
	}
	return nil
}
