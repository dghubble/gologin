package google

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
	google "google.golang.org/api/oauth2/v2"
)

func TestGoogleHandler(t *testing.T) {
	jsonData := `{"id": "900913", "name": "Ben Bitdiddle"}`
	expectedUser := &google.Userinfoplus{Id: "900913", Name: "Ben Bitdiddle"}
	proxyClient, server := newGoogleTestServer(jsonData)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		googleUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, googleUser)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// GoogleHandler assert that:
	// - Token is read from the ctx and passed to the Google API
	// - google Userinfoplus is obtained from the Google API
	// - success handler is called
	// - google Userinfoplus is added to the ctx of the success handler
	googleHandler := googleHandler(config, ctxh.ContextHandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	googleHandler.ServeHTTP(ctx, w, req)
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestGoogleHandler_MissingCtxToken(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Context missing Token", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// GoogleHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	googleHandler := googleHandler(config, success, ctxh.ContextHandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	googleHandler.ServeHTTP(context.Background(), w, req)
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
	failure := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, ErrUnableToGetGoogleUser, err)
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// GoogleHandler cannot get Google User, assert that:
	// - failure handler is called
	// - error cannot get Google User added to the failure handler ctx
	googleHandler := googleHandler(config, success, ctxh.ContextHandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	googleHandler.ServeHTTP(ctx, w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestValidateResponse(t *testing.T) {
	assert.Equal(t, nil, validateResponse(&google.Userinfoplus{Id: "123"}, nil))
	assert.Equal(t, ErrUnableToGetGoogleUser, validateResponse(nil, fmt.Errorf("Server error")))
	assert.Equal(t, ErrCannotValidateGoogleUser, validateResponse(nil, nil))
	assert.Equal(t, ErrCannotValidateGoogleUser, validateResponse(&google.Userinfoplus{Name: "Ben"}, nil))
}
