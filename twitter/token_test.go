package twitter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/logintest"
)

func TestTokenHandler_successEndToEnd(t *testing.T) {
	proxyClient, _, server := newTwitterTestServer(testTwitterUserJSON)
	defer server.Close()
	// returns an http.Client which proxies requests to Twitter test server
	proxyClientSource := logintest.NewStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(successChecks(t)),
		Failure:      gologin.ErrorHandlerFunc(logintest.ErrorOnFailure(t)),
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST token to server under test
	resp, err := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestTokenHandler_wrongMethod(t *testing.T) {
	proxyClient, _, server := newTwitterTestServer(testTwitterUserJSON)
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
	proxyClient, _, server := newTwitterTestServer(testTwitterUserJSON)
	defer server.Close()
	proxyClientSource := logintest.NewStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      gologin.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST token to server under test
	resp, _ := http.PostForm(ts.URL, nil)
	logintest.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{"wrongFieldName": {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, "wrongFieldName": {testTwitterTokenSecret}})
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
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetTwitterUser.Error()+"\n")
}

func TestTokenHandler_whenValidationServerDown(t *testing.T) {
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
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetTwitterUser.Error()+"\n")
}
