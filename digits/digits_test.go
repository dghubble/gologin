package digits

import (
	"errors"
	"net/http"
	"testing"

	"github.com/dghubble/go-digits/digits"
)

func TestValidateResponse(t *testing.T) {
	emptyAccount := new(digits.Account)
	validAccount := &digits.Account{
		AccessToken: digits.AccessToken{Token: "token", Secret: "secret"},
	}
	successResp := &http.Response{
		StatusCode: 200,
	}
	badResp := &http.Response{
		StatusCode: 400,
	}
	respErr := errors.New("some error decoding Account")

	// success case
	if err := validateResponse(validAccount, successResp, nil); err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	// error cases
	errorCases := []error{
		// account missing credentials
		validateResponse(emptyAccount, successResp, nil),
		// Digits account API did not return a 200
		validateResponse(validAccount, badResp, nil),
		// Network error or JSON unmarshalling error
		validateResponse(validAccount, successResp, respErr),
		validateResponse(validAccount, badResp, respErr),
	}
	for _, err := range errorCases {
		if err != ErrUnableToGetDigitsAccount {
			t.Errorf("expected %v, got %v", ErrUnableToGetDigitsAccount, err)
		}
	}
}
