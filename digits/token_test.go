package digits

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dghubble/oauth1"
	"github.com/jbcjorge/gologin"
	oauth1Login "github.com/jbcjorge/gologin/oauth1"
	"github.com/jbcjorge/gologin/testutils"
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

	// oauth1 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth1.HTTPClient, proxyClient)

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
		fmt.Fprintf(w, "success handler called")
	}

	// TokenHandler assert that:
	// - access token/secret are read from POST
	// - digits account is obtained from the Digits accounts endpoint
	// - success handler is called
	// - digits account is added to the success handler ctx
	tokenHandler := TokenHandler(config, http.HandlerFunc(success), testutils.AssertFailureNotCalled(t))
	w := httptest.NewRecorder()
	form := url.Values{accessTokenField: {testDigitsToken}, accessTokenSecretField: {testDigitsSecret}}
	req, _ := http.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	tokenHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestTokenHandler_ErrorVerifyingToken(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Digits Account Endpoint Down", http.StatusInternalServerError)
	defer server.Close()

	// oauth1 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth1.HTTPClient, proxyClient)

	config := &oauth1.Config{}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.Error(t, err) {
			assert.Equal(t, err, ErrUnableToGetDigitsAccount)
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// TokenHandler cannot verify Digits account, assert that:
	// - failure handler is called
	// - error is added to the failure handler context
	tokenHandler := TokenHandler(config, testutils.AssertSuccessNotCalled(t), http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	form := url.Values{accessTokenField: {testDigitsToken}, accessTokenSecretField: {testDigitsSecret}}
	req, _ := http.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	tokenHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
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
