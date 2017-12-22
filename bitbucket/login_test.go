package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jbcjorge/gologin"
	oauth2Login "github.com/jbcjorge/gologin/oauth2"
	"github.com/jbcjorge/gologin/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestBitbucketHandler(t *testing.T) {
	jsonData := `{"username": "bitster", "display_name": "Atlas Ian"}`
	expectedUser := &User{Username: "bitster", DisplayName: "Atlas Ian"}
	proxyClient, server := newBitbucketTestServer(jsonData)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		bitbucketUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, bitbucketUser)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// BitbucketHandler assert that:
	// - Token is read from the ctx and passed to the Bitbucket API
	// - bitbucket User is obtained from the Bitbucket API
	// - success handler is called
	// - bitbucket User is added to the ctx of the success handler
	bitbucketHandler := bitbucketHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	bitbucketHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestBitbucketHandler_MissingCtxToken(t *testing.T) {
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

	// BitbucketHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	bitbucketHandler := bitbucketHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	bitbucketHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestBitbucketHandler_ErrorGettingUser(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Bitbucket Service Down", http.StatusInternalServerError)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, ErrUnableToGetBitbucketUser, err)
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// BitbucketHandler cannot get Bitbucket User, assert that:
	// - failure handler is called
	// - error cannot get Bitbucket User added to the failure handler ctx
	bitbucketHandler := bitbucketHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	bitbucketHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestValidateResponse(t *testing.T) {
	validUser := &User{Username: "bitster"}
	validResponse := &http.Response{StatusCode: 200}
	invalidResponse := &http.Response{StatusCode: 500}
	assert.Equal(t, nil, validateResponse(validUser, validResponse, nil))
	assert.Equal(t, ErrUnableToGetBitbucketUser, validateResponse(validUser, validResponse, fmt.Errorf("Server error")))
	assert.Equal(t, ErrUnableToGetBitbucketUser, validateResponse(validUser, invalidResponse, nil))
	assert.Equal(t, ErrUnableToGetBitbucketUser, validateResponse(&User{}, validResponse, nil))
}
