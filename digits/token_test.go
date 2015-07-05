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
	testDigitsToken       = "some-token"
	testDigitsTokenSecret = "some-token-secret"
)

func TestValidateToken_missingToken(t *testing.T) {
	err := validateToken("", testDigitsTokenSecret)
	if err != ErrMissingToken {
		t.Errorf("expected error %v, got %v", ErrMissingToken, err)
	}
}

func TestValidateToken_missingTokenSecret(t *testing.T) {
	err := validateToken(testDigitsToken, "")
	if err != ErrMissingTokenSecret {
		t.Errorf("expected error %v, got %v", ErrMissingTokenSecret, err)
	}
}

func TestTokenHandler_successEndToEnd(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()
	proxyClientSource := logintest.NewStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		// returns an http.Client which proxies requests to the Digits test server
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(successChecks(t)),
		Failure:      gologin.ErrorHandlerFunc(logintest.ErrorOnFailure(t)),
	}
	// server under test
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST Digits access token
	resp, err := http.PostForm(ts.URL, url.Values{accessTokenField: {testDigitsToken}, accessTokenSecretField: {testDigitsTokenSecret}})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestTokenHandler_wrongMethod(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()
	proxyClientSource := logintest.NewStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	resp, _ := http.Get(ts.URL)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected response code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestTokenHandler_invalidPOSTFields(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()
	proxyClientSource := logintest.NewStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST Digits access Token
	resp, _ := http.PostForm(ts.URL, url.Values{"wrongFieldName": {testDigitsToken}, accessTokenSecretField: {testDigitsTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{accessTokenField: {testDigitsToken}, "wrongFieldName": {testDigitsTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrMissingTokenSecret.Error()+"\n")
}

func TestTokenHandler_unauthorized(t *testing.T) {
	proxyClient, server := logintest.UnauthorizedTestServer()
	defer server.Close()
	proxyClientSource := logintest.NewStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST Digits access Token
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testDigitsToken}, accessTokenSecretField: {testDigitsTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestTokenHandler_digitsAPIDown(t *testing.T) {
	client, _, server := logintest.TestServer()
	defer server.Close()
	// source returns client to a NoOp server
	proxyClientSource := logintest.NewStubClientSource(client)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))

	// POST Digits Access Token
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testDigitsToken}, accessTokenSecretField: {testDigitsTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}
