package oauth2

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

// LoginHandler

func TestLoginHandler(t *testing.T) {
	expectedState := "state_val"
	expectedRedirect := "https://api.example.com/authorize?client_id=client_id&redirect_uri=redirect_url&response_type=code&state=state_val"
	config := &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "redirect_url",
		Endpoint: oauth2.Endpoint{
			AuthURL: "https://api.example.com/authorize",
		},
	}
	failure := testutils.AssertFailureNotCalled(t)

	// LoginHandler assert that:
	// - redirects to the oauth2.Config AuthURL
	// - redirect status code is 302
	// - redirect url is the OAuth2 Config RedirectURL with the ClientID and ctx state
	loginHandler := LoginHandler(config, failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := WithState(context.Background(), expectedState)
	loginHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, expectedRedirect, w.HeaderMap.Get("Location"))
}

func TestLoginHandler_MissingCtxState(t *testing.T) {
	config := &oauth2.Config{}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Context missing state value", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// LoginHandler cannot get the state from the ctx, assert that:
	// - failure handler is called
	// - error about missing state is added to the ctx
	loginHandler := LoginHandler(config, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	loginHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

// CallbackHandler

func TestCallbackHandler(t *testing.T) {
	jsonData := `{
       "access_token":"2YotnFZFEjr1zCsicMWpAA",
       "token_type":"example",
       "refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
       "example_parameter":"example_value"
     }`
	expectedToken := &oauth2.Token{
		AccessToken:  "2YotnFZFEjr1zCsicMWpAA",
		TokenType:    "example",
		RefreshToken: "tGzv3JOkF0XG5Qx2TlKWIA",
	}
	server := NewAccessTokenServer(t, jsonData)
	defer server.Close()

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			TokenURL: server.URL,
		},
	}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := TokenFromContext(ctx)
		assert.Equal(t, expectedToken.AccessToken, token.AccessToken)
		assert.Equal(t, expectedToken.TokenType, token.Type())
		assert.Equal(t, expectedToken.RefreshToken, token.RefreshToken)
		// real oauth2.Token populates internal raw field and unmockable Expiry time
		assert.Nil(t, err)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// CallbackHandler gets OAuth2 access token, assert that:
	// - success handler is called
	// - access token added to the ctx of the success handler
	// - failure handler is not called
	callbackHandler := CallbackHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?code=any_code&state=d4e5f6", nil)
	ctx := WithState(context.Background(), "d4e5f6")
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestCallbackHandler_ParseCallbackError(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Request missing code or state", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler called without code or state, assert that:
	// - failure handler is called
	// - error about missing code or state is added to the ctx
	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?code=any_code", nil)
	callbackHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/?state=any_state", nil)
	callbackHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestCallbackHandler_MissingCtxState(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Context missing state value", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler called without state param in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing state is added to the failure handler ctx
	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?code=any_code&state=d4e5f6", nil)
	callbackHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestCallbackHandler_StateMismatch(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Invalid OAuth2 state parameter", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler ctx state does not match state param, assert that:
	// - failure handler is called
	// - error about invalid state param is added to the failure handler ctx
	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?code=any_code&state=d4e5f6", nil)
	ctx := WithState(context.Background(), "differentState")
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestCallbackHandler_ExchangeError(t *testing.T) {
	_, server := testutils.NewErrorServer("OAuth2 Service Down", http.StatusInternalServerError)
	defer server.Close()

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			TokenURL: server.URL,
		},
	}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			// error from golang.org/x/oauth2 config.Exchange as provider is down
			assert.True(t, strings.HasPrefix(err.Error(), "oauth2: cannot fetch token"))
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// CallbackHandler cannot exchange for an Access Token, assert that:
	// - failure handler is called
	// - error with the reason the exchange failed is added to the ctx
	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?code=any_code&state=d4e5f6", nil)
	ctx := WithState(context.Background(), "d4e5f6")
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}
