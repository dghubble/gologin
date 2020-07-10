package oauth1

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/testutils"
	"github.com/dghubble/oauth1"
	"github.com/stretchr/testify/assert"
)

// LoginHandler

func TestLoginHandler(t *testing.T) {
	expectedToken := "request_token"
	expectedSecret := "request_secret"
	data := url.Values{}
	data.Add("oauth_token", expectedToken)
	data.Add("oauth_token_secret", expectedSecret)
	data.Add("oauth_callback_confirmed", "true")
	server := NewRequestTokenServer(t, data)
	defer server.Close()

	config := &oauth1.Config{
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: server.URL,
		},
	}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		requestToken, requestSecret, err := RequestTokenFromContext(ctx)
		assert.Equal(t, expectedToken, requestToken)
		assert.Equal(t, expectedSecret, requestSecret)
		assert.Nil(t, err)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// LoginHandler gets OAuth1 request token, assert that:
	// - success handler is called
	// - request token added to the ctx of the success handler
	// - request secret added to the ctx of the success handler
	// - failure handler is not called
	loginHandler := LoginHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	loginHandler.ServeHTTP(w, req)
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestLoginHandler_RequestTokenError(t *testing.T) {
	_, server := testutils.NewErrorServer("OAuth1 Server Error", http.StatusInternalServerError)
	defer server.Close()

	config := &oauth1.Config{
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: server.URL,
		},
	}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			// first validation in OAuth1 impl failed
			assert.Equal(t, "oauth1: Server returned status 500", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// LoginHandler cannot get the OAuth1 request token, assert that:
	// - failure handler is called
	// - error about StatusInternalServerError is added to the ctx
	loginHandler := LoginHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	loginHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

// AuthRedirectHandler

func TestAuthRedirectHandler(t *testing.T) {
	requestToken := "request_token"
	expectedRedirect := "https://api.example.com/authorize?oauth_token=request_token"
	config := &oauth1.Config{
		Endpoint: oauth1.Endpoint{
			AuthorizeURL: "https://api.example.com/authorize",
		},
	}
	failure := testutils.AssertFailureNotCalled(t)

	// AuthRedirectHandler redirects to the AuthorizationURL, assert that:
	// - redirect status code is 302
	// - redirect url is the OAuth1 AuthorizeURL with the ctx request token
	authRedirectHandler := AuthRedirectHandler(config, failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := WithRequestToken(context.Background(), requestToken, "")
	authRedirectHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, expectedRedirect, w.HeaderMap.Get("Location"))
}

func TestAuthRedirectHandler_MissingCtxRequestToken(t *testing.T) {
	config := &oauth1.Config{}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth1: Context missing request token or secret", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler cannot get the request token from the ctx, assert that:
	// - failure handler is called
	// - error about missing request token is added to the ctx
	authRedirectHandler := AuthRedirectHandler(config, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	authRedirectHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestAuthRedirectHandler_AuthorizationURL(t *testing.T) {
	requestToken := "request_token"
	config := &oauth1.Config{
		Endpoint: oauth1.Endpoint{
			AuthorizeURL: "%gh&%ij", // always causes AuthorizationURL parse error
		},
	}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "parse %gh&%ij: invalid URL escape \"%gh\"", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// AuthRedirectHandler cannot construct the AuthorizationURL, assert that:
	// - failure handler is called
	// - error about authorization URL is added to the ctx
	authRedirectHandler := AuthRedirectHandler(config, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := WithRequestToken(context.Background(), requestToken, "")
	authRedirectHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}

// CallbackHandler

func TestCallbackHandler(t *testing.T) {
	expectedToken := "acces_token"
	expectedSecret := "access_secret"
	requestSecret := "request_secret"
	data := url.Values{}
	data.Add("oauth_token", expectedToken)
	data.Add("oauth_token_secret", expectedSecret)
	server := NewAccessTokenServer(t, data)
	defer server.Close()

	config := &oauth1.Config{
		Endpoint: oauth1.Endpoint{
			AccessTokenURL: server.URL,
		},
	}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		accessToken, accessSecret, err := AccessTokenFromContext(ctx)
		assert.Equal(t, expectedToken, accessToken)
		assert.Equal(t, expectedSecret, accessSecret)
		assert.Nil(t, err)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// CallbackHandler gets OAuth1 access token, assert that:
	// - success handler is called
	// - access token and secret added to the ctx of the success handler
	// - failure handler is not called
	callbackHandler := CallbackHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?oauth_token=any_token&oauth_verifier=any_verifier", nil)
	ctx := WithRequestToken(context.Background(), "", requestSecret)
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestCallbackHandler_ParseAuthorizationCallbackError(t *testing.T) {
	config := &oauth1.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth1: Request missing oauth_token or oauth_verifier", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler called without oauth_token or oauth_verifier, assert that:
	// - failure handler is called
	// - error about missing oauth_token or oauth_verifier is added to the ctx
	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?oauth_verifier=", nil)
	callbackHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestCallbackHandler_MissingCtxRequestSecret(t *testing.T) {
	config := &oauth1.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth1: Context missing request token or secret", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler cannot get the request secret from the ctx, assert that:
	// - failure handler is called
	// - error about missing request secret is added to the ctx
	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?oauth_token=any_token&oauth_verifier=any_verifier", nil)
	callbackHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestCallbackHandler_AccessTokenError(t *testing.T) {
	requestSecret := "request_secret"
	_, server := testutils.NewErrorServer("OAuth1 Server Error", http.StatusInternalServerError)
	defer server.Close()

	config := &oauth1.Config{
		Endpoint: oauth1.Endpoint{
			AccessTokenURL: server.URL,
		},
	}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			// first validation in OAuth1 impl failed
			assert.Equal(t, "oauth1: Server returned status 500", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler cannot get the OAuth1 access token, assert that:
	// - failure handler is called
	// - error about StatusInternalServerError is added to the ctx
	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?oauth_token=any_token&oauth_verifier=any_verifier", nil)
	ctx := WithRequestToken(context.Background(), "", requestSecret)
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}
