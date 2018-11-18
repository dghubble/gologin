package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/gologin/testutils"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestGithubHandler(t *testing.T) {
	jsonData := `{"id": 917408, "name": "Alyssa Hacker"}`
	expectedUser := &github.User{ID: github.Int64(917408), Name: github.String("Alyssa Hacker")}
	proxyClient, server := newGithubTestServer(jsonData)
	defer server.Close()

	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	config.Endpoint.AuthURL = "https://github.com/login/oauth/authorize"
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		githubUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, githubUser)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// GithubHandler assert that:
	// - Token is read from the ctx and passed to the Github API
	// - github User is obtained from the Github API
	// - success handler is called
	// - github User is added to the ctx of the success handler
	githubHandler := githubHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	githubHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestGithubHandler_MissingCtxToken(t *testing.T) {
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

	// GithubHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	githubHandler := githubHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	githubHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestGithubHandler_ErrorGettingUser(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Github Service Down", http.StatusInternalServerError)
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
			assert.Equal(t, ErrUnableToGetGithubUser, err)
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// GithubHandler cannot get Github User, assert that:
	// - failure handler is called
	// - error cannot get Github User added to the failure handler ctx
	githubHandler := githubHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	githubHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestValidateResponse(t *testing.T) {
	validUser := &github.User{ID: github.Int64(123)}
	validResponse := &github.Response{Response: &http.Response{StatusCode: 200}}
	invalidResponse := &github.Response{Response: &http.Response{StatusCode: 500}}
	assert.Equal(t, nil, validateResponse(validUser, validResponse, nil))
	assert.Equal(t, ErrUnableToGetGithubUser, validateResponse(validUser, validResponse, fmt.Errorf("Server error")))
	assert.Equal(t, ErrUnableToGetGithubUser, validateResponse(validUser, invalidResponse, nil))
	assert.Equal(t, ErrUnableToGetGithubUser, validateResponse(&github.User{}, validResponse, nil))
}

func Test_githubClientFromAuthURL(t *testing.T) {
	for _, test := range []struct {
		authURL          string
		expClientBaseURL string
	}{
		{authURL: "https://github.com/login/oauth/authorize/", expClientBaseURL: "https://api.github.com/"},
		{authURL: "https://github.com/login/oauth/authorize", expClientBaseURL: "https://api.github.com/"},
		{authURL: "https://github.mycompany.com/login/oauth/authorize", expClientBaseURL: "https://github.mycompany.com/api/v3/"},
		{authURL: "http://github.mycompany.com/login/oauth/authorize", expClientBaseURL: "http://github.mycompany.com/api/v3/"},
	} {
		client, err := githubClientFromAuthURL(test.authURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := client.BaseURL.String(), test.expClientBaseURL; got != want {
			t.Errorf("For authorization URL %q, expected client URL %q, but got %q", test.authURL, want, got)
		}
	}
}
