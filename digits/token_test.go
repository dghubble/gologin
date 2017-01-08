package digits

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/gologin/testutils"
	"github.com/dghubble/oauth1"
	"github.com/stretchr/testify/assert"
)

func TestValidateToken_missingToken(t *testing.T) {
	err := validateToken("", testDigitsSecret)
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

func TestTokenHandler(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()

	config := &oauth1.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		account, err := AccountFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testDigitsToken, account.AccessToken.Token)
		assert.Equal(t, testDigitsSecret, account.AccessToken.Secret)
		assert.Equal(t, "0123456789", account.PhoneNumber)

		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testDigitsToken, accessToken)
		assert.Equal(t, testDigitsSecret, accessSecret)
	}
	handler := TokenHandler(config, http.HandlerFunc(success), testutils.AssertFailureNotCalled(t))
	// oauth1 Client will use the proxy client's base Transport
	ts := httptest.NewServer(oauth1Login.WithHTTPClient(proxyClient, handler))
	// POST Digits access token to server under test
	resp, err := http.PostForm(ts.URL, url.Values{accessTokenField: {testDigitsToken}, accessTokenSecretField: {testDigitsSecret}})
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	}
}

func TestTokenHandler_ErrorVerifyingToken(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Digits Account Endpoint Down", http.StatusInternalServerError)
	defer server.Close()

	config := &oauth1.Config{}
	handler := TokenHandler(config, testutils.AssertSuccessNotCalled(t), nil)
	// oauth1 Client will use the proxy client's base Transport
	ts := httptest.NewServer(oauth1Login.WithHTTPClient(proxyClient, handler))
	// assert that error occurs indicating the Digits Account could not be confirmed
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testDigitsToken}, accessTokenSecretField: {testDigitsSecret}})
	testutils.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestTokenHandler_NonPost(t *testing.T) {
	config := &oauth1.Config{}
	handler := TokenHandler(config, testutils.AssertSuccessNotCalled(t), nil)
	ts := httptest.NewServer(handler)
	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	// assert that default (nil) failure handler returns a 405 Method Not Allowed
	if assert.NotNil(t, resp) {
		// TODO: change to 405
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestTokenHandler_InvalidFields(t *testing.T) {
	config := &oauth1.Config{}
	handler := TokenHandler(config, testutils.AssertSuccessNotCalled(t), nil)
	ts := httptest.NewServer(handler)

	// asert errors occur for different missing POST fields
	resp, err := http.PostForm(ts.URL, url.Values{"wrongFieldName": {testDigitsToken}, accessTokenSecretField: {testDigitsSecret}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")

	resp, err = http.PostForm(ts.URL, url.Values{accessTokenField: {testDigitsToken}, "wrongFieldName": {testDigitsSecret}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrMissingTokenSecret.Error()+"\n")
}
