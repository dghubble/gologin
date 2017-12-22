package digits

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/go-digits/digits"
	"github.com/jbcjorge/gologin/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	testConsumerKey          = "mykey"
	testAccountEndpoint      = "https://api.digits.com/1.1/sdk/account.json"
	testAccountRequestHeader = `OAuth oauth_consumer_key="mykey",`
	testAccountJSON          = `{"access_token": {"token": "some-token", "secret": "some-secret"}, "phone_number": "0123456789"}`
	testDigitsToken          = "some-token"
	testDigitsSecret         = "some-secret"
)

func TestValidateEcho_missingAccountEndpoint(t *testing.T) {
	err := validateEcho("", testAccountRequestHeader, testConsumerKey)
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountEndpoint, err)
	}
}

func TestValidateEcho_missingAccountRequestHeader(t *testing.T) {
	err := validateEcho(testAccountEndpoint, "", testConsumerKey)
	if assert.Error(t, err) {
		assert.Equal(t, ErrMissingAccountRequestHeader, err)
	}
}

func TestValidateEcho_digitsEndpoint(t *testing.T) {
	cases := []struct {
		endpoint string
		valid    bool
	}{
		{"https://api.digits.com/1.1/sdk/account.json", true},
		{"http://api.digits.com/1.1/sdk/account.json", false},
		{"https://digits.com/1.1/sdk/account.json", false},
		{"https://evil.com/1.1/sdk/account.json", false},
		// respect the path defined in Digits javascript sdk
		{"https://api.digits.com/2.0/future/so/cool.json", true},
	}
	for _, c := range cases {
		err := validateEcho(c.endpoint, testAccountRequestHeader, testConsumerKey)
		if c.valid {
			assert.Nil(t, err)
		} else {
			if assert.Error(t, err) {
				assert.Equal(t, ErrInvalidDigitsEndpoint, err)
			}
		}
	}
}

func TestValidateEcho_headerConsumerKey(t *testing.T) {
	cases := []struct {
		header string
		valid  bool
	}{
		{`OAuth oauth_consumer_key="mykey"`, true},
		// wrong consumer key
		{`OAuth oauth_consumer_key="wrongkey"`, false},
		// empty consumer key
		{`OAuth oauth_consumer_key=""`, false},
		// missing value quotes
		{`OAuth oauth_consumer_key=mykey`, false},
		// no oauth_consumer_key field
		{`OAuth oauth_token="mykey"`, false},
		{"OAuth", false},
	}
	for _, c := range cases {
		err := validateEcho(testAccountEndpoint, c.header, testConsumerKey)
		if c.valid {
			assert.Nil(t, err)
		} else {
			if assert.Error(t, err) {
				assert.Equal(t, ErrInvalidConsumerKey, err)
			}
		}
	}
}

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

func TestWebHandler(t *testing.T) {
	proxyClient, _, server := newDigitsTestServer(testAccountJSON)
	defer server.Close()

	config := &Config{
		ConsumerKey: testConsumerKey,
		Client:      proxyClient,
	}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		account, err := AccountFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testDigitsToken, account.AccessToken.Token)
		assert.Equal(t, testDigitsSecret, account.AccessToken.Secret)
		assert.Equal(t, "0123456789", account.PhoneNumber)

		endpoint, header, err := EchoFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testAccountEndpoint, endpoint)
		assert.Equal(t, testAccountRequestHeader, header)
	}
	handler := LoginHandler(config, http.HandlerFunc(success), testutils.AssertFailureNotCalled(t))
	ts := httptest.NewServer(handler)
	// POST OAuth Echo to server under test
	resp, err := http.PostForm(ts.URL, url.Values{accountEndpointField: {testAccountEndpoint}, accountRequestHeaderField: {testAccountRequestHeader}})
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestWebHandler_ErrorGettingRequestToken(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("OAuth1 Service Down", http.StatusInternalServerError)
	defer server.Close()

	config := &Config{
		ConsumerKey: testConsumerKey,
		Client:      proxyClient,
	}
	handler := LoginHandler(config, testutils.AssertSuccessNotCalled(t), nil)
	ts := httptest.NewServer(handler)
	// assert that error occurs indicating the Digits Account cound not be confirmed
	resp, _ := http.PostForm(ts.URL, url.Values{accountEndpointField: {testAccountEndpoint}, accountRequestHeaderField: {testAccountRequestHeader}})
	testutils.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func TestWebHandler_NonPost(t *testing.T) {
	config := &Config{}
	handler := LoginHandler(config, testutils.AssertSuccessNotCalled(t), nil)
	ts := httptest.NewServer(handler)
	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	// assert that default (nil) failure handler returns a 405 Method Not Allowed
	if assert.NotNil(t, resp) {
		// TODO: change to 405
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestWebHandler_InvalidFields(t *testing.T) {
	config := &Config{
		ConsumerKey: testConsumerKey,
	}
	handler := LoginHandler(config, testutils.AssertSuccessNotCalled(t), nil)
	ts := httptest.NewServer(handler)

	// assert errors occur for different missing/incorrect POST fields
	resp, err := http.PostForm(ts.URL, url.Values{"wrongKeyName": {testAccountEndpoint}, accountRequestHeaderField: {testAccountRequestHeader}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrMissingAccountEndpoint.Error()+"\n")

	resp, err = http.PostForm(ts.URL, url.Values{accountEndpointField: {"https://evil.com"}, accountRequestHeaderField: {testAccountRequestHeader}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrInvalidDigitsEndpoint.Error()+"\n")

	resp, err = http.PostForm(ts.URL, url.Values{accountEndpointField: {testAccountEndpoint}, accountRequestHeaderField: {`OAuth oauth_consumer_key="notmyconsumerkey",`}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrInvalidConsumerKey.Error()+"\n")

	// valid, but incorrect Digits account endpoint
	resp, err = http.PostForm(ts.URL, url.Values{accountEndpointField: {"https://api.digits.com/1.1/wrong.json"}, accountRequestHeaderField: {testAccountRequestHeader}})
	assert.Nil(t, err)
	testutils.AssertBodyString(t, resp.Body, ErrUnableToGetDigitsAccount.Error()+"\n")
}

func checkSuccess(t *testing.T) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		account, err := AccountFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testDigitsToken, account.AccessToken.Token)
		assert.Equal(t, testDigitsSecret, account.AccessToken.Secret)
		assert.Equal(t, "0123456789", account.PhoneNumber)

		endpoint, header, err := EchoFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, testAccountEndpoint, endpoint)
		assert.Equal(t, testAccountRequestHeader, header)
	}
	return http.HandlerFunc(fn)
}
