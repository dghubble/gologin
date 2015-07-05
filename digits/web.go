package digits

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dghubble/go-digits/digits"
	"github.com/dghubble/gologin"
	"github.com/dghubble/sling"
)

const (
	accountEndpointField      = "accountEndpoint"
	accountRequestHeaderField = "accountRequestHeader"
)

// Errors for missing echo data, invalid data, and errors gettting Digits
// accounts.
var (
	ErrMissingAccountEndpoint      = fmt.Errorf("digits: missing OAuth Echo form field %s", accountEndpointField)
	ErrMissingAccountRequestHeader = fmt.Errorf("digits: missing OAuth Echo form field %s", accountRequestHeaderField)
	ErrInvalidDigitsEndpoint       = errors.New("digits: invalid Digits endpoint")
	ErrInvalidConsumerKey          = errors.New("digits: incorrect OAuth Echo Auth Header Consumer Key")
	consumerKeyRegexp              = regexp.MustCompile("oauth_consumer_key=\"(.*?)\"")
)

// LoginHandlerConfig configures a LoginHandler.
type LoginHandlerConfig struct {
	ConsumerKey string
	HTTPClient  *http.Client
	Success     SuccessHandler
	Failure     gologin.ErrorHandler
}

// LoginHandler handles Digits OAuth1 Echo login requests. If echo data
// validates and authentication succeeds, handling is delegated to a
// SuccessHandler which is provided with the Digits Account. Otherwise, an
// ErrorHandler handles responding.
type LoginHandler struct {
	consumerKey string
	httpClient  *http.Client
	success     SuccessHandler
	failure     gologin.ErrorHandler
}

// NewLoginHandler returns a new LoginHandler.
func NewLoginHandler(config *LoginHandlerConfig) *LoginHandler {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	failure := config.Failure
	if failure == nil {
		failure = gologin.DefaultErrorHandler
	}
	return &LoginHandler{
		consumerKey: config.ConsumerKey,
		httpClient:  httpClient,
		success:     config.Success,
		failure:     failure,
	}
}

// ServeHTTP receives POSTed Digits OAuth Echo headers, validates them, and
// fetches the Digits Account. If successful, handling is delegated to the
// SuccessHandler. Otherwise, the ErrorHandler is called.
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		h.failure.ServeHTTP(w, nil, http.StatusMethodNotAllowed)
		return
	}
	req.ParseForm()
	accountEndpoint := req.PostForm.Get(accountEndpointField)
	accountRequestHeader := req.PostForm.Get(accountRequestHeaderField)
	// validate POST'ed Digits OAuth Echo data
	err := validateEcho(accountEndpoint, accountRequestHeader, h.consumerKey)
	if err != nil {
		h.failure.ServeHTTP(w, err, http.StatusBadRequest)
		return
	}
	// fetch Digits Account
	account, resp, err := requestAccount(h.httpClient, accountEndpoint, accountRequestHeader)
	// validate the Digits Account
	err = validateResponse(account, resp, err)
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
