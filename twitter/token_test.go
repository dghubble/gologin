package twitter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dghubble/gologin/v2"
	oauth1Login "github.com/dghubble/gologin/v2/oauth1"
	"github.com/dghubble/gologin/v2/testutils"
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

	// oauth1 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth1.HTTPClient, proxyClient)

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
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// TokenHandler assert that:
	// - access token/secret are read from POST
	// - twitter User is obtained from Twitter API
	// - success handler is called
	// - twitter User is added to success handler ctx
	tokenHandler := TokenHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	form := url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}}
	req, _ := http.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	tokenHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestTokenHandler_ErrorVerifyingToken(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Twitter Verify Credentials Down", http.StatusInternalServerError)
	defer server.Close()

	// oauth1 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth1.HTTPClient, proxyClient)

	config := &oauth1.Config{}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.Error(t, err) {
			assert.Equal(t, err, ErrUnableToGetTwitterUser)
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// TokenHandler cannot verify Twitter credentials and get User, assert that:
	// - failure handler is called
	// - error is added to the failure handler context
	tokenHandler := TokenHandler(config, testutils.AssertSuccessNotCalled(t), http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	form := url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}}
	req, _ := http.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	tokenHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
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
