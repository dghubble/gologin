package digits

import (
	"net/http"
	"testing"

	"github.com/dghubble/go-digits/digits"
)

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
