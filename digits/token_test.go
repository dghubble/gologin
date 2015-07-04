package login

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		// returns an http.Client which proxies requests to the digits test server
		AuthConfig: proxyClientSource,
		Success:    SuccessHandlerFunc(successChecks(t)),
		Failure:    ErrorHandlerFunc(errorOnFailure(t)),
	}
	// server under test
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST Digits access token
	resp, err := http.PostForm(ts.URL, url.Values{"digitsToken": {testDigitsToken}, "digitsTokenSecret": {testDigitsTokenSecret}})
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
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		AuthConfig: proxyClientSource,
		Success:    SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:    DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	resp, _ := http.Get(ts.URL)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected response code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestTokenHandler_invalidPOSTArguments(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		AuthConfig: proxyClientSource,
		Success:    SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:    DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST Digits access Token
	resp, _ := http.PostForm(ts.URL, url.Values{"wrongFieldName": {testDigitsToken}, "digitsTokenSecret": {testDigitsTokenSecret}})
	assertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")
	resp, _ = http.PostForm(ts.URL, url.Values{"digitsToken": {testDigitsToken}, "wrongFieldName": {testDigitsTokenSecret}})
	assertBodyString(t, resp.Body, ErrMissingTokenSecret.Error()+"\n")
}

func TestTokenHandler_unauthorized(t *testing.T) {
	proxyClient, _, server := newRejectingTestServer()
	defer server.Close()
	proxyClientSource := newStubClientSource(proxyClient)

	handlerConfig := &TokenHandlerConfig{
		AuthConfig: proxyClientSource,
		Success:    SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:    DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))
	// POST Digits access Token
	resp, _ := http.PostForm(ts.URL, url.Values{"digitsToken": {testDigitsToken}, "digitsTokenSecret": {testDigitsTokenSecret}})
	assertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestTokenHandler_digitsAPIDown(t *testing.T) {
	client, _, server := testServer()
	defer server.Close()
	// source returns client to a NoOp server
	proxyClientSource := newStubClientSource(client)

	handlerConfig := &TokenHandlerConfig{
		AuthConfig: proxyClientSource,
		Success:    SuccessHandlerFunc(errorOnSuccess(t)),
		Failure:    DefaultErrorHandler,
	}
	ts := httptest.NewServer(NewTokenHandler(handlerConfig))

	// POST Digits Access Token
	resp, _ := http.PostForm(ts.URL, url.Values{"digitsToken": {testDigitsToken}, "digitsTokenSecret": {testDigitsTokenSecret}})
	assertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
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
