package digits

import (
	"fmt"
	"net/http"

	"github.com/dghubble/go-digits/digits"
	"github.com/dghubble/gologin"
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

// TokenHandler receives a POSTed Digits token/secret and verifies the Digits
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

// ServeHTTP receives a POSTed Digits token/secret and verifies the Digits
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
	// verify/lookup the Digits Account
	httpClient := h.oauth1Config.GetClient(accessToken, accessTokenSecret)
	digitsClient := digits.NewClient(httpClient)
	account, resp, err := digitsClient.Accounts.Account()
	err = validateResponse(account, resp, err)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	h.success.ServeHTTP(w, req, account)
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
