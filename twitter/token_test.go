package twitter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	login "github.com/dghubble/go-login"
	"github.com/dghubble/go-login/logintest"
)

func TestTokenHandler_successEndToEnd(t *testing.T) {
	proxyClient, _, server := newTwitterTestServer(testTwitterUserJSON)
	defer server.Close()
	// returns an http.Client which proxies requests to Twitter test server
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(successChecks(t)),
		Failure:      login.ErrorHandlerFunc(logintest.ErrorOnFailure(t)),
	}
	// server under test, which uses go-login/twitter TokenHandler
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
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      login.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	resp, _ := http.Get(ts.URL)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected response code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestTokenHandler_invalidPOSTField(t *testing.T) {
	proxyClient, _, server := newTwitterTestServer(testTwitterUserJSON)
	defer server.Close()
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      login.DefaultErrorHandler,
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
	proxyClient, _, server := newRejectingTestServer()
	defer server.Close()
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      login.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetTwitterUser.Error()+"\n")
}

func TestTokenHandler_whenValidationServerDown(t *testing.T) {
	client, _, server := logintest.TestServer()
	defer server.Close()
	// source returns client to a NoOp server
	proxyClientSource := newStubClientSource(client)

	handlerConfig := &TokenHandlerConfig{
		OAuth1Config: proxyClientSource,
		Success:      SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:      login.DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetTwitterUser.Error()+"\n")
}

type stubClientSource struct {
	client *http.Client
}

// newStubClientSource returns a stubClientSource which always returns the
// given http client.
func newStubClientSource(client *http.Client) *stubClientSource {
	return &stubClientSource{
		client: client,
	}
}

func (s *stubClientSource) GetClient(token, tokenSecret string) *http.Client {
	return s.client
}
