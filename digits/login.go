package digits

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dghubble/go-digits/digits"
	"github.com/dghubble/sling"
	"github.com/jbcjorge/gologin"
)

const (
	accountEndpointField      = "accountEndpoint"
	accountRequestHeaderField = "accountRequestHeader"
)

// Digits login errors
var (
	ErrUnableToGetDigitsAccount    = fmt.Errorf("digits: unable to get Digits account")
	ErrMissingAccountEndpoint      = fmt.Errorf("digits: missing OAuth Echo endpoint field")
	ErrMissingAccountRequestHeader = fmt.Errorf("digits: missing OAuth Echo header field")
	ErrInvalidDigitsEndpoint       = fmt.Errorf("digits: invalid Digits endpoint")
	ErrInvalidConsumerKey          = fmt.Errorf("digits: incorrect OAuth Echo Auth Header Consumer Key")
	consumerKeyRegexp              = regexp.MustCompile("oauth_consumer_key=\"(.*?)\"")
)

// Config configures Digits Handlers.
type Config struct {
	// Digits Consumer Key required to verify the OAuth Echo response.
	ConsumerKey string
	// Client to use to make the Accounts Endpoint request. If nil, then
	// http.DefaultClient is used.
	Client *http.Client
}

// LoginHandler receives a Digits OAuth Echo endpoint and OAuth header,
// validates the echo, and calls the endpoint to get the corresponding Digits
// Account. If successful, the Digits Account is added to the ctx and the
// success handler is called. Otherwise, the failure handler is called.
func LoginHandler(config *Config, success, failure http.Handler) http.Handler {
	success = getAccountViaEcho(config, success, failure)
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
		accountEndpoint := req.PostForm.Get(accountEndpointField)
		accountRequestHeader := req.PostForm.Get(accountRequestHeaderField)
		// validate POST'ed Digits OAuth Echo data
		err := validateEcho(accountEndpoint, accountRequestHeader, config.ConsumerKey)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithEcho(ctx, accountEndpoint, accountRequestHeader)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// getAccountViaEcho is a http.Handler that gets the Digits Echo endpoint and
// OAuth header from the ctx and calls the endpoint to get the corresponding
// Digits Account. If successful, the Account is added to the ctx and the
// success handler is called. Otherwise, the failure handler is called.
func getAccountViaEcho(config *Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	client := config.Client
	if client == nil {
		client = http.DefaultClient
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		endpoint, header, err := EchoFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		// fetch the Digits Account
		account, resp, err := requestAccount(client, endpoint, header)
		// validate the Digits Account response
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

// validateResponse returns an error if the given Digits Account, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(account *digits.Account, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK || account == nil {
		return ErrUnableToGetDigitsAccount
	}
	if token := account.AccessToken; token.Token == "" || token.Secret == "" {
		// JSON deserialized Digits account is missing fields
		return ErrUnableToGetDigitsAccount
	}
	return nil
}
