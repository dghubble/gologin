package apple

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/gologin/v2"
	oauth2Login "github.com/dghubble/gologin/v2/oauth2"
	"github.com/dghubble/gologin/v2/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestAppleHandler(t *testing.T) {
	setupAppleTestKeys()

	expectedUser := &User{
		ID:    "000929.993e01a8ec174de3b28b806530c17c26.2444",
		Email: "ivy@harvard.edu",
	}

	proxyClient, _, server := testutils.TestServer()
	defer server.Close()

	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	ctx = oauth2Login.WithToken(ctx, buildTokenWithOIDC(expectedUser.ID, expectedUser.Email, "testkey"))

	config := &oauth2.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		appleUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, appleUser)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// AppleHandler assert that:
	// - success handler is called
	// - apple User is added to the ctx of the success handler
	appleHandler := appleHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	appleHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestAppleHandler_MissingCtxToken(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Context missing Token", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// AppleHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	appleHandler := appleHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	appleHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestAppleLoginHandler(t *testing.T) {
	expectedState := "state_val"
	expectedRedirect := AppleBaseURL + "/auth/authorize?client_id=client_id&redirect_uri=redirect_url&response_mode=form_post&response_type=code&state=state_val"
	config := &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "redirect_url",
		Endpoint: oauth2.Endpoint{
			AuthURL: AppleBaseURL + "/auth/authorize",
		},
	}
	failure := testutils.AssertFailureNotCalled(t)

	// LoginHandler assert that:
	// - redirects to the oauth2.Config AuthURL
	// - redirect status code is 302
	// - redirect url is the OAuth2 Config RedirectURL with the ClientID and ctx state and response_mode=form_post which Apple requires
	loginHandler := appleLoginHandler(config, failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := oauth2Login.WithState(context.Background(), expectedState)
	loginHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, expectedRedirect, w.HeaderMap.Get("Location"))
}
