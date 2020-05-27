package google

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
	google "google.golang.org/api/oauth2/v2"
)

func TestGoogleHandler(t *testing.T) {
	jsonData := `{"id": "900913", "name": "Ben Bitdiddle"}`
	expectedUser := &google.Userinfo{Id: "900913", Name: "Ben Bitdiddle"}
	proxyClient, server := newGoogleTestServer(jsonData)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		googleUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		// assert required fields; Userinfoplus contains other raw response info
		assert.Equal(t, expectedUser.Id, googleUser.Id)
		assert.Equal(t, expectedUser.Id, googleUser.Id)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// GoogleHandler assert that:
	// - Token is read from the ctx and passed to the Google API
	// - google Userinfoplus is obtained from the Google API
	// - success handler is called
	// - google Userinfoplus is added to the ctx of the success handler
	googleHandler := googleHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	googleHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestGoogleHandler_MissingCtxToken(t *testing.T) {
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

	// GoogleHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	googleHandler := googleHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	googleHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestGoogleHandler_ErrorGettingUser(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Google Service Down", http.StatusInternalServerError)
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
			assert.Equal(t, ErrUnableToGetGoogleUser, err)
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// GoogleHandler cannot get Google User, assert that:
	// - failure handler is called
	// - error cannot get Google User added to the failure handler ctx
	googleHandler := googleHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	googleHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestValidateResponse(t *testing.T) {
	assert.Equal(t, nil, validateResponse(&google.Userinfo{Id: "123"}, nil))
	assert.Equal(t, ErrUnableToGetGoogleUser, validateResponse(nil, fmt.Errorf("Server error")))
	assert.Equal(t, ErrCannotValidateGoogleUser, validateResponse(nil, nil))
	assert.Equal(t, ErrCannotValidateGoogleUser, validateResponse(&google.Userinfo{Name: "Ben"}, nil))
}
