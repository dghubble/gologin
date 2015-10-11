package facebook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/ctxh"
	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/gologin/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func TestFacebookHandler(t *testing.T) {
	jsonData := `{"id": "54638001", "name": "Ben Bitdiddle"}`
	expectedFacebookUser := &User{ID: "54638001", Name: "Ben Bitdiddle"}
	proxyClient, server := newFacebookTestServer(jsonData)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		facebookUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedFacebookUser, facebookUser)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// FacebookHandler assert that:
	// - Token is read from the ctx and passed to the facebook API
	// - facebook User is obtained from the facebook API
	// - success handler is called
	// - facebook User is added to the ctx of the success handler
	facebookHandler := facebookHandler(config, ctxh.ContextHandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	facebookHandler.ServeHTTP(ctx, w, req)
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestFacebookHandler_MissingCtxToken(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Context missing Token", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// FacebookHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	facebookHandler := facebookHandler(config, success, ctxh.ContextHandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	facebookHandler.ServeHTTP(context.Background(), w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestFacebookHandler_ErrorGettingUser(t *testing.T) {
	proxyClient, server := testutils.ErrorServer(t, "Facebook Service Down", http.StatusInternalServerError)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, ErrUnableToGetFacebookUser, err)
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// FacebookHandler cannot get Facebook User, assert that:
	// - failure handler is called
	// - error cannot get Facebook User added to the failure handler ctx
	facebookHandler := facebookHandler(config, success, ctxh.ContextHandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	facebookHandler.ServeHTTP(ctx, w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}
