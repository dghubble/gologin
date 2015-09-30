package oauth1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	contentType     = "Content-Type"
	formContentType = "application/x-www-form-urlencoded"
)

// NewRequestTokenServer returns a new httptest.Server OAuth1 provider Request
// Token endpoint. Caller must close the server.
func NewRequestTokenServer(t *testing.T, data url.Values) *httptest.Server {
	return NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.NotEmpty(t, req.Header.Get("Authorization"))
		w.Header().Set(contentType, formContentType)
		w.Write([]byte(data.Encode()))
	})
}

// NewErrorServer returns a new httptest.Server OAuth1 provider endpoint which
// responds with the given error message and code. Caller must close the
// server.
func NewErrorServer(t *testing.T, message string, code int) *httptest.Server {
	return NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.NotEmpty(t, req.Header.Get("Authorization"))
		http.Error(w, message, code)
	})
}

// NewAccessTokenServer creates a new httptest.Server OAuth1 provider Access
// Token endpoint. Caller must close the server.
func NewAccessTokenServer(t *testing.T, data url.Values) *httptest.Server {
	return NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.NotEmpty(t, req.Header.Get("Authorization"))
		w.Header().Set(contentType, formContentType)
		w.Write([]byte(data.Encode()))
	})
}

// NewTestServeFunc is an adapter to allow the use of ordinary functions as
// httptest.Server's for testing. Caller must close the server.
func NewTestServerFunc(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}
