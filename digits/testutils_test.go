package digits

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/go-digits/digits"
	"github.com/dghubble/gologin/logintest"
)

func newDigitsTestServer(jsonData string) (*http.Client, *http.ServeMux, *httptest.Server) {
	client, mux, server := logintest.TestServer()
	mux.HandleFunc("/1.1/sdk/account.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, mux, server
}

// successChecks is a SuccessHandler which checks that the test Digits Account
// was passed.
func successChecks(t *testing.T) func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
	success := func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
		if account.AccessToken.Token != "t" {
			t.Errorf("expected Token value t, got %q", account.AccessToken.Token)
		}
		if account.AccessToken.Secret != "s" {
			t.Errorf("expected Secret value s, got %q", account.AccessToken.Secret)
		}
		if account.PhoneNumber != "0123456789" {
			t.Errorf("expected PhoneNumber 0123456789, got %q", account.PhoneNumber)
		}
	}
	return success
}

func errorOnSuccess(t *testing.T) func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
	success := func(w http.ResponseWriter, req *http.Request, account *digits.Account) {
		t.Errorf("unexpected call to success, %v", account)
	}
	return success
}
