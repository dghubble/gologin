package login

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/go-digits/digits"
)

const (
	testConsumerKey          = "mykey"
	testAccountEndpoint      = "https://api.digits.com/1.1/sdk/account.json"
	testAccountRequestHeader = `OAuth oauth_consumer_key="mykey",`
	testAccountJSON          = `{"access_token": {"token": "t", "secret": "s"}, "phone_number": "0123456789"}`
)

func TestValidateEcho_missingAccountEndpoint(t *testing.T) {
	err := validateEcho("", testAccountRequestHeader, testConsumerKey)
	if err != ErrMissingAccountEndpoint {
		t.Errorf("expected error %v, got %v", ErrMissingAccountEndpoint, err)
	}
}

func TestValidateEcho_missingAccountRequestHeader(t *testing.T) {
	err := validateEcho(testAccountEndpoint, "", testConsumerKey)
	if err != ErrMissingAccountRequestHeader {
		t.Errorf("expected error %v, got %v", ErrMissingAccountRequestHeader, err)
	}
}

func TestValidateEcho_digitsEndpoint(t *testing.T) {
	cases := []struct {
		endpoint string
		valid    bool
	}{
		{"https://api.digits.com/1.1/sdk/account.json", true},
		{"http://api.digits.com/1.1/sdk/account.json", false},
		{"https://digits.com/1.1/sdk/account.json", false},
		{"https://evil.com/1.1/sdk/account.json", false},
		// respect the path defined in Digits javascript sdk
		{"https://api.digits.com/2.0/future/so/cool.json", true},
	}
	for _, c := range cases {
		err := validateEcho(c.endpoint, testAccountRequestHeader, testConsumerKey)
		if c.valid && err != nil {
			t.Errorf("expected endpoint %q to be valid, got error %v", c.endpoint, err)
		}
		if !c.valid && err != ErrInvalidDigitsEndpoint {
			t.Errorf("expected endpoint %q to be invalid, got error %v", c.endpoint, err)
		}
	}
}

func TestValidateEcho_headerConsumerKey(t *testing.T) {
	cases := []struct {
		header string
		valid  bool
	}{
		{`OAuth oauth_consumer_key="mykey"`, true},
		// wrong consumer key
		{`OAuth oauth_consumer_key="wrongkey"`, false},
		// empty consumer key
		{`OAuth oauth_consumer_key=""`, false},
		// missing value quotes
		{`OAuth oauth_consumer_key=mykey`, false},
		// no oauth_consumer_key field
		{`OAuth oauth_token="mykey"`, false},
		{"OAuth", false},
	}
	for _, c := range cases {
		err := validateEcho(testAccountEndpoint, c.header, testConsumerKey)
		if c.valid && err != nil {
			t.Errorf("expected header %q to be valid, got error %v", c.header, err)
		}
		if !c.valid && err != ErrInvalidConsumerKey {
			t.Errorf("expected header %q to be invalid, got error %v", c.header, err)
		}
	}
}

func TestValidateAccountResponse(t *testing.T) {
	emptyAccount := new(digits.Account)
	validAccount := &digits.Account{
		AccessToken: digits.AccessToken{Token: "token", Secret: "secret"},
	}
	successResp := &http.Response{
		StatusCode: 200,
	}
	badResp := &http.Response{
		StatusCode: 400,
	}
	respErr := errors.New("some error decoding Account")

	// success case
	if err := validateAccountResponse(validAccount, successResp, nil); err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	// error cases
	errorCases := []error{
		// account missing credentials
		validateAccountResponse(emptyAccount, successResp, nil),
		// Digits account API did not return a 200
		validateAccountResponse(validAccount, badResp, nil),
		// Network error or JSON unmarshalling error
		validateAccountResponse(validAccount, successResp, respErr),
		validateAccountResponse(validAccount, badResp, respErr),
	}
	for _, err := range errorCases {
		if err != ErrUnableToGetDigitsAccount {
			t.Errorf("expected %v, got %v", ErrUnableToGetDigitsAccount, err)
		}
	}
}

func TestErrorHandler(t *testing.T) {
	const expectedMessage = "digits: some error"
	rec := httptest.NewRecorder()
	// should pass through errors and codes
	DefaultErrorHandler.ServeHTTP(rec, fmt.Errorf(expectedMessage), http.StatusBadRequest)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected code %v, got %v", http.StatusBadRequest, rec.Code)
	}
	if rec.Body.String() != expectedMessage+"\n" {
		t.Errorf("expected error message %v, got %v", expectedMessage+"\n", rec.Body.String())
	}
}

func TestWebHandler_successEndToEnd(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()

	handlerConfig := &WebHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(successChecks(t)),
		Failure:     ErrorHandlerFunc(errorOnFailure(t)),
	}
	// setup server under test
	ts := httptest.NewServer(NewWebHandler(handlerConfig))
	// POST OAuth Echo data
	resp, err := http.PostForm(ts.URL, url.Values{"accountEndpoint": {testAccountEndpoint}, "accountRequestHeader": {testAccountRequestHeader}})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestWebHandler_wrongMethod(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()

	handlerConfig := &WebHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewWebHandler(handlerConfig))
	resp, _ := http.Get(ts.URL)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected response code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestWebHandler_invalidPOSTArguments(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()

	handlerConfig := &WebHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewWebHandler(handlerConfig))
	// POST Digits OAuth Echo data
	resp, _ := http.PostForm(ts.URL, url.Values{"wrongKeyName": {testAccountEndpoint}, "accountRequestHeader": {testAccountRequestHeader}})
	assertBodyString(t, resp.Body, ErrMissingAccountEndpoint.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{"accountEndpoint": {"https://evil.com"}, "accountRequestHeader": {testAccountRequestHeader}})
	assertBodyString(t, resp.Body, ErrInvalidDigitsEndpoint.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{"accountEndpoint": {testAccountEndpoint}, "accountRequestHeader": {`OAuth oauth_consumer_key="notmyconsumerkey",`}})
	assertBodyString(t, resp.Body, ErrInvalidConsumerKey.Error()+"\n")
	// valid, but incorrect Digits account endpoint
	resp, _ = http.PostForm(ts.URL, url.Values{"accountEndpoint": {"https://api.digits.com/1.1/wrong.json"}, "accountRequestHeader": {testAccountRequestHeader}})
	assertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestWebHandler_unauthorized(t *testing.T) {
	proxyClient, _, server := newRejectingTestServer()
	defer server.Close()

	handlerConfig := &WebHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewWebHandler(handlerConfig))
	// POST OAuth Echo data
	resp, _ := http.PostForm(ts.URL, url.Values{"accountEndpoint": {testAccountEndpoint}, "accountRequestHeader": {testAccountRequestHeader}})
	assertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestWebHandler_digitsAPIDown(t *testing.T) {
	// NoOp server
	client, _, server := testServer()
	defer server.Close()

	handlerConfig := &WebHandlerConfig{
		HTTPClient:  client,
		ConsumerKey: testConsumerKey,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewWebHandler(handlerConfig))
	// POST OAuth Echo data
	resp, _ := http.PostForm(ts.URL, url.Values{"accountEndpoint": {testAccountEndpoint}, "accountRequestHeader": {testAccountRequestHeader}})
	assertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

// success and failure handlers for testing

func successChecks(t *testing.T) func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
	success := func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
		if account.AccessToken.Token != "t" {
			t.Errorf("expected Token value t, got %q", account.AccessToken.Token)
		}
		if account.AccessToken.Secret != "s" {
			t.Errorf("expected Secret value s, got %q", account.AccessToken.Secret)
		}
		if account.PhoneNumber != "0123456789" {
			t.Errorf("expected PhoneNumber 0123456789, got %q", account.PhoneNumber)
		}
	}
	return success
}

func errorOnSuccess(t *testing.T) func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
	success := func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
		t.Errorf("unexpected call to success, %v", account)
	}
	return success
}

func errorOnFailure(t *testing.T) func(w http.ResponseWriter, err error, code int) {
	failure := func(w http.ResponseWriter, err error, code int) {
		t.Errorf("unexpected call to failure, %v %d", err, code)
	}
	return failure
}
