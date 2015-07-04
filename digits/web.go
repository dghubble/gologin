package login

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dghubble/go-digits/digits"
	"github.com/dghubble/sling"
)

const (
	accountEndpointKey      = "accountEndpoint"
	accountRequestHeaderKey = "accountRequestHeader"
)

// Errors for missing echo data, invalid data, and errors gettting Digits
// accounts.
var (
	ErrMissingAccountEndpoint      = fmt.Errorf("digits: missing OAuth Echo form field %s", accountEndpointKey)
	ErrMissingAccountRequestHeader = fmt.Errorf("digits: missing OAuth Echo form field %s", accountRequestHeaderKey)
	ErrInvalidDigitsEndpoint       = errors.New("digits: invalid Digits endpoint")
	ErrInvalidConsumerKey          = errors.New("digits: incorrect OAuth Echo Auth Header Consumer Key")
	ErrUnableToGetDigitsAccount    = errors.New("digits: unable to get Digits account")
	consumerKeyRegexp              = regexp.MustCompile("oauth_consumer_key=\"(.*?)\"")
)

// WebHandlerConfig configures a WebHandler.
type WebHandlerConfig struct {
	ConsumerKey string
	HTTPClient  *http.Client
	Success     SuccessHandler
	Failure     ErrorHandler
}

// WebHandler receives POSTed Web OAuth Echo headers, validates them, and
// fetches the Digits Account. If successful, handling is delegated to the
// SuccessHandler. Otherwise, the ErrorHandler is called.
type WebHandler struct {
	consumerKey string
	httpClient  *http.Client
	success     SuccessHandler
	failure     ErrorHandler
}

// NewWebHandler returns a new WebHandler.
func NewWebHandler(config *WebHandlerConfig) *WebHandler {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &WebHandler{
		consumerKey: config.ConsumerKey,
		httpClient:  httpClient,
		success:     config.Success,
		failure:     config.Failure,
	}
}

// ServeHTTP receives POSTed Web OAuth Echo headers, validates them, and
// fetches the Digits Account. If successful, handling is delegated to the
// SuccessHandler. Otherwise, the ErrorHandler is called.
func (h *WebHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		h.failure.ServeHTTP(w, nil, http.StatusMethodNotAllowed)
		return
	}
	req.ParseForm()
	accountEndpoint := req.PostForm.Get(accountEndpointKey)
	accountRequestHeader := req.PostForm.Get(accountRequestHeaderKey)
	// validate POST'ed Digits OAuth Echo data
	err := validateEcho(accountEndpoint, accountRequestHeader, h.consumerKey)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	// fetch Digits Account
	account, resp, err := requestAccount(h.httpClient, accountEndpoint, accountRequestHeader)
	// validate the Digits Account
	err = validateAccountResponse(account, resp, err)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	h.success.ServeHTTP(w, req, account)
}

// requestAccount makes a request to the Digits account endpoint using the
// provided Authorization header.
func requestAccount(client *http.Client, accountEndpoint, authorizationHeader string) (*digits.Account, *http.Response, error) {
	request, err := http.NewRequest("GET", accountEndpoint, nil)
	if err != nil {
		return nil, nil, ErrInvalidDigitsEndpoint
	}
	request.Header.Set("Authorization", authorizationHeader)
	account := new(digits.Account)
	resp, err := sling.New().Client(client).Do(request, account, nil)
	return account, resp, err
}

// validateEcho checks that the Digits OAuth Echo arguments are valid. If the
// endpoint does not match the Digits API or the header does not include the
// correct consumer key, a non-nil error is returned.
func validateEcho(accountEndpoint, accountRequestHeader, consumerKey string) error {
	if accountEndpoint == "" {
		return ErrMissingAccountEndpoint
	}
	if accountRequestHeader == "" {
		return ErrMissingAccountRequestHeader
	}
	// check accountEndpoint matches expected protocol/domain
	if !strings.HasPrefix(accountEndpoint, digits.DigitsAPI) {
		return ErrInvalidDigitsEndpoint
	}
	// validate the OAuth Echo data's auth header consumer key
	matches := consumerKeyRegexp.FindStringSubmatch(accountRequestHeader)
	if len(matches) != 2 || matches[1] != consumerKey {
		return ErrInvalidConsumerKey
	}
	return nil
}

// validateAccountResponse returns an error if the given Digits Account, raw
// http.Response, or error from Digits are unexpected. Returns nil if the
// account response is valid.
func validateAccountResponse(account *digits.Account, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK || account == nil {
		return ErrUnableToGetDigitsAccount
	}
	if token := account.AccessToken; token.Token == "" || token.Secret == "" {
		// JSON deserialized Digits account is missing fields
		return ErrUnableToGetDigitsAccount
	}
	return nil
}

// SuccessHandler is called when authentication via Digits succeeds.
type SuccessHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, account *digits.Account)
}

// ErrorHandler is called when authentication via Digits fails.
type ErrorHandler interface {
	ServeHTTP(w http.ResponseWriter, err error, code int)
}

// DefaultErrorHandler responds to requests by passing through the error
// message and code from the login library.
var DefaultErrorHandler = &passthroughErrorHandler{}

type passthroughErrorHandler struct{}

func (e passthroughErrorHandler) ServeHTTP(w http.ResponseWriter, err error, code int) {
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	http.Error(w, "", code)
}

// HandlerFunc adapters

// SuccessHandlerFunc is an adapter to allow an ordinary function to be used as
// a SuccessHandler.
type SuccessHandlerFunc func(w http.ResponseWriter, req *http.Request, account *digits.Account)

func (f SuccessHandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, account *digits.Account) {
	f(w, req, account)
}

// ErrorHandlerFunc is an adapter to allow an ordinary function to be used as
// an ErrorHandlerFunc.
type ErrorHandlerFunc func(w http.ResponseWriter, err error, code int)

func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, err error, code int) {
	f(w, err, code)
}
