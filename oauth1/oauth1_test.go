package oauth1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dghubble/gologin/v2/testutils"
	"github.com/stretchr/testify/assert"
)

const (
	contentType     = "Content-Type"
	formContentType = "application/x-www-form-urlencoded"
)

// NewRequestTokenServer returns a new httptest.Server OAuth1 provider Request
// Token endpoint. Caller must close the server.
func NewRequestTokenServer(t *testing.T, data url.Values) *httptest.Server {
	return testutils.NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.NotEmpty(t, req.Header.Get("Authorization"))
		w.Header().Set(contentType, formContentType)
		w.Write([]byte(data.Encode()))
	})
}

// NewAccessTokenServer creates a new httptest.Server OAuth1 provider Access
// Token endpoint. Caller must close the server.
func NewAccessTokenServer(t *testing.T, data url.Values) *httptest.Server {
	return testutils.NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.NotEmpty(t, req.Header.Get("Authorization"))
		w.Header().Set(contentType, formContentType)
		w.Write([]byte(data.Encode()))
	})
}
