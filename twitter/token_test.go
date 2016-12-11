package twitter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/gologin"
	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/gologin/testutils"
	"github.com/dghubble/oauth1"
	"github.com/stretchr/testify/assert"
)

const (
	testTwitterToken             = "some-token"
	testTwitterTokenSecret       = "some-secret"
	testTwitterUserJSON          = `{"id": 1234, "id_str": "1234", "screen_name": "gopher"}`
	expectedUserID         int64 = 1234
)

func TestTokenHandler(t *testing.T) {
	proxyClient, _, server := newTwitterVerifyServer(testTwitterUserJSON)
	defer server.Close()

	config := &oauth1.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testTwitterToken, accessToken)
		assert.Equal(t, testTwitterTokenSecret, accessSecret)

		user, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUserID, user.ID)
		assert.Equal(t, "1234", user.IDStr)
	}
	handler := TokenHandler(config, http.HandlerFunc(success), testutils.AssertFailureNotCalled(t))
	// oauth1 Client will use the proxy client's base Transport
	ts := httptest.NewServer(oauth1Login.WithHTTPClient(proxyClient, handler))
	// POST token to server under test
	resp, err := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	}
}

func TestTokenHandler_ErrorVerifyingToken(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Twitter Verify Credentials Down", http.StatusInternalServerError)
	defer server.Close()

	config := &oauth1.Config{}
	handler := TokenHandler(config, testutils.AssertSuccessNotCalled(t), nil)
	// oauth1 Client will use the proxy client's base Transport
	ts := httptest.NewServer(oauth1Login.WithHTTPClient(proxyClient, handler))
	// assert that error occurs indicating the Twitter User could not be confirmed
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	testutils.AssertBodyString(t, resp.Body, ErrUnableToGetTwitterUser.Error()+"\n")
}

func TestTokenHandler_ErrorVerifyingTokenPassesError(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Twitter Verify Credentials Down", http.StatusInternalServerError)
	defer server.Close()

	config := &oauth1.Config{}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		// assert that error passed through ctx
		err := gologin.ErrorFromContext(ctx)
		if assert.Error(t, err) {
			assert.Equal(t, err, ErrUnableToGetTwitterUser)
		}
	}
	handler := TokenHandler(config, testutils.AssertSuccessNotCalled(t), http.HandlerFunc(failure))
	// oauth1 Client will use the proxy client's base Transport
	ts := httptest.NewServer(oauth1Login.WithHTTPClient(proxyClient, handler))
	http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
}

func TestTokenHandler_NonPost(t *testing.T) {
	config := &oauth1.Config{}
	ts := httptest.NewServer(TokenHandler(config, testutils.AssertSuccessNotCalled(t), nil))
	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	// assert that default (nil) failure handler returns a 405 Method Not Allowed
	if assert.NotNil(t, resp) {
		// TODO: change to 405
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestTokenHandler_NonPostPassesError(t *testing.T) {
	config := &oauth1.Config{}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		// assert that Method not allowed error passed through ctx
		err := gologin.ErrorFromContext(ctx)
		if assert.Error(t, err) {
			assert.Equal(t, err, fmt.Errorf("Method not allowed"))
		}
	}
	ts := httptest.NewServer(TokenHandler(config, testutils.AssertSuccessNotCalled(t), http.HandlerFunc(failure)))
	http.Get(ts.URL)
}

func TestTokenHandler_InvalidFields(t *testing.T) {
	config := &oauth1.Config{}
	ts := httptest.NewServer(TokenHandler(config, testutils.AssertSuccessNotCalled(t), nil))

	// assert errors occur for different missing POST fields
	resp, err := http.PostForm(ts.URL, nil)
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")

	resp, err = http.PostForm(ts.URL, url.Values{"wrongFieldName": {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")

	resp, err = http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, "wrongFieldName": {testTwitterTokenSecret}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrMissingTokenSecret.Error()+"\n")
}
