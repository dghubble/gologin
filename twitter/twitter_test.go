package twitter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/go-login/logintest"
	"github.com/dghubble/go-twitter/twitter"
)

const (
	testTwitterUserJSON = `{"id": 1234, "id_str": "1234", "screen_name": "somename"}`
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

// successChecks is a SuccessHandler which checks that the testTwitterUserJSON
// User was passed.
func successChecks(t *testing.T) func(w http.ResponseWriter, req *http.Request, user *twitter.User) {
	return func(w http.ResponseWriter, req *http.Request, user *twitter.User) {
		if user.ID != 1234 || user.IDStr != "1234" {
			t.Errorf("expected SuccessHandler to receive Twitter User, got %+v", user)
		}
		return
	}
}

func errorOnSuccess(t *testing.T) func(w http.ResponseWriter, req *http.Request, user *twitter.User) {
	return func(w http.ResponseWriter, req *http.Request, user *twitter.User) {
		t.Errorf("unexpected SuccessHandler call with User %v", user)
	}
}
