package twitter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/gologin/logintest"
)

const (
	testTwitterToken       = "some-token"
	testTwitterTokenSecret = "some-token-secret"
	testTwitterUserJSON    = `{"id": 1234, "id_str": "1234", "screen_name": "somename"}`
)

func newTwitterTestServer(jsonData string) (*http.Client, *http.ServeMux, *httptest.Server) {
	client, mux, server := logintest.TestServer()
	mux.HandleFunc("/1.1/account/verify_credentials.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, mux, server
}

func newRejectingTestServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	client, mux, server := logintest.TestServer()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	return client, mux, server
}

// successChecks is a SuccessHandler which checks that the test Twitter User
// and test token/secret were passed.
func successChecks(t *testing.T) func(w http.ResponseWriter, req *http.Request, user *twitter.User, token, tokenSecret string) {
	return func(w http.ResponseWriter, req *http.Request, user *twitter.User, token, tokenSecret string) {
		if user.ID != 1234 || user.IDStr != "1234" {
			t.Errorf("expected SuccessHandler to receive Twitter User, got %+v", user)
		}
		if token != testTwitterToken {
			t.Errorf("expected SuccessHandler to receive token %v, got %v", testTwitterToken, token)
		}
		if tokenSecret != testTwitterTokenSecret {
			t.Errorf("expected SuccessHandler to receive token secret %v, got %v", testTwitterTokenSecret, tokenSecret)
		}
		return
	}
}

func errorOnSuccess(t *testing.T) func(w http.ResponseWriter, req *http.Request, user *twitter.User, token, tokenSecret string) {
	return func(w http.ResponseWriter, req *http.Request, user *twitter.User, token, tokenSecret string) {
		t.Errorf("unexpected SuccessHandler call")
	}
}
