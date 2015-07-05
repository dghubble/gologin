package digits

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/logintest"
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

func TestWebHandler_successEndToEnd(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()

	handlerConfig := &LoginHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(successChecks(t)),
		Failure:     gologin.ErrorHandlerFunc(logintest.ErrorOnFailure(t)),
	}
	// setup server under test
	ts := httptest.NewServer(NewLoginHandler(handlerConfig))
	// POST OAuth Echo data
	resp, err := http.PostForm(ts.URL, url.Values{accountEndpointField: {testAccountEndpoint}, accountRequestHeaderField: {testAccountRequestHeader}})
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

	handlerConfig := &LoginHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewLoginHandler(handlerConfig))
	resp, _ := http.Get(ts.URL)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected response code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestWebHandler_invalidPOSTArguments(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()

	handlerConfig := &LoginHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewLoginHandler(handlerConfig))
	// POST Digits OAuth Echo data
	resp, _ := http.PostForm(ts.URL, url.Values{"wrongKeyName": {testAccountEndpoint}, accountRequestHeaderField: {testAccountRequestHeader}})
	logintest.AssertBodyString(t, resp.Body, ErrMissingAccountEndpoint.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{accountEndpointField: {"https://evil.com"}, accountRequestHeaderField: {testAccountRequestHeader}})
	logintest.AssertBodyString(t, resp.Body, ErrInvalidDigitsEndpoint.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{accountEndpointField: {testAccountEndpoint}, accountRequestHeaderField: {`OAuth oauth_consumer_key="notmyconsumerkey",`}})
	logintest.AssertBodyString(t, resp.Body, ErrInvalidConsumerKey.Error()+"\n")
	// valid, but incorrect Digits account endpoint
	resp, _ = http.PostForm(ts.URL, url.Values{accountEndpointField: {"https://api.digits.com/1.1/wrong.json"}, accountRequestHeaderField: {testAccountRequestHeader}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestWebHandler_unauthorized(t *testing.T) {
	proxyClient, server := logintest.UnauthorizedTestServer()
	defer server.Close()

	handlerConfig := &LoginHandlerConfig{
		ConsumerKey: testConsumerKey,
		HTTPClient:  proxyClient,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewLoginHandler(handlerConfig))
	// POST OAuth Echo data
	resp, _ := http.PostForm(ts.URL, url.Values{accountEndpointField: {testAccountEndpoint}, accountRequestHeaderField: {testAccountRequestHeader}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestWebHandler_digitsAPIDown(t *testing.T) {
	client, _, server := logintest.TestServer()
	defer server.Close()

	handlerConfig := &LoginHandlerConfig{
		HTTPClient:  client,
		ConsumerKey: testConsumerKey,
		Success:     SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:     gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewLoginHandler(handlerConfig))
	// POST OAuth Echo data
	resp, _ := http.PostForm(ts.URL, url.Values{accountEndpointField: {testAccountEndpoint}, accountRequestHeaderField: {testAccountRequestHeader}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}
