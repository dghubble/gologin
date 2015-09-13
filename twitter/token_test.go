package twitter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/logintest"
	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/oauth1"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const (
	testTwitterToken             = "some-token"
	testTwitterTokenSecret       = "some-secret"
	testTwitterUserJSON          = `{"id": 1234, "id_str": "1234", "screen_name": "gopher"}`
	expectedUserID         int64 = 1234
)

func TestTokenHandler(t *testing.T) {
	proxyClient, _, server := newTwitterTestServer(testTwitterUserJSON)
	defer server.Close()
	// oauth1 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth1.HTTPClient, proxyClient)

	config := &oauth1.Config{}
	success := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		accessToken, accessSecret, err := oauth1Login.AccessTokenFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testTwitterToken, accessToken)
		assert.Equal(t, testTwitterTokenSecret, accessSecret)

		user, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUserID, user.ID)
		assert.Equal(t, "1234", user.IDStr)
	}
	handler := TokenHandler(config, ctxh.ContextHandlerFunc(success), assertFailureNotCalled(t))
	ts := httptest.NewServer(ctxh.NewHandlerWithContext(ctx, handler))
	// POST token to server under test
	resp, err := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, resp.StatusCode, http.StatusOK)
	}
}

func TestTokenHandler_Unauthorized(t *testing.T) {
	proxyClient, server := logintest.UnauthorizedTestServer()
	defer server.Close()
	// oauth1 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth1.HTTPClient, proxyClient)

	config := &oauth1.Config{}
	handler := TokenHandler(config, assertSuccessNotCalled(t), nil)
	ts := httptest.NewServer(ctxh.NewHandlerWithContext(ctx, handler))
	// assert that error occurs indicating the Twitter User could not be confirmed
	resp, _ := http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	logintest.AssertBodyString(t, resp.Body, ErrUnableToGetTwitterUser.Error()+"\n")
}

func TestTokenHandler_UnauthorizedPassesError(t *testing.T) {
	proxyClient, server := logintest.UnauthorizedTestServer()
	defer server.Close()
	// oauth1 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth1.HTTPClient, proxyClient)

	config := &oauth1.Config{}
	failure := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		// assert that error passed through ctx
		err := gologin.ErrorFromContext(ctx)
		if assert.Error(t, err) {
			assert.Equal(t, err, ErrUnableToGetTwitterUser)
		}
	}
	handler := TokenHandler(config, assertSuccessNotCalled(t), ctxh.ContextHandlerFunc(failure))
	ts := httptest.NewServer(ctxh.NewHandlerWithContext(ctx, handler))
	http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
}

func TestTokenHandler_NonPost(t *testing.T) {
	config := &oauth1.Config{}
	ts := httptest.NewServer(ctxh.NewHandler(TokenHandler(config, assertSuccessNotCalled(t), nil)))
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
	failure := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		// assert that Method not allowed error passed through ctx
		err := gologin.ErrorFromContext(ctx)
		if assert.Error(t, err) {
			assert.Equal(t, err, fmt.Errorf("Method not allowed"))
		}
	}
	ts := httptest.NewServer(ctxh.NewHandler(TokenHandler(config, assertSuccessNotCalled(t), ctxh.ContextHandlerFunc(failure))))
	http.Get(ts.URL)
}

func TestTokenHandler_InvalidFields(t *testing.T) {
	config := &oauth1.Config{}
	ts := httptest.NewServer(ctxh.NewHandler(TokenHandler(config, assertSuccessNotCalled(t), nil)))

	// assert errors occur for different missing POST fields
	resp, err := http.PostForm(ts.URL, nil)
	assert.Nil(t, err)
	logintest.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")

	resp, err = http.PostForm(ts.URL, url.Values{"wrongFieldName": {testTwitterToken}, accessTokenSecretField: {testTwitterTokenSecret}})
	assert.Nil(t, err)
	logintest.AssertBodyString(t, resp.Body, ErrMissingToken.Error()+"\n")

	resp, err = http.PostForm(ts.URL, url.Values{accessTokenField: {testTwitterToken}, "wrongFieldName": {testTwitterTokenSecret}})
	assert.Nil(t, err)
	logintest.AssertBodyString(t, resp.Body, ErrMissingTokenSecret.Error()+"\n")
}

func newTwitterTestServer(jsonData string) (*http.Client, *http.ServeMux, *httptest.Server) {
	client, mux, server := logintest.TestServer()
	mux.HandleFunc("/1.1/account/verify_credentials.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, mux, server
}

func assertSuccessNotCalled(t *testing.T) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to success ContextHandler")
	}
	return ctxh.ContextHandlerFunc(fn)
}

func assertFailureNotCalled(t *testing.T) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to failure ContextHandler")
	}
	return ctxh.ContextHandlerFunc(fn)
}
