package oauth2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	contentType     = "Content-Type"
	jsonContentType = "application/json"
)

// NewAccessTokenServer creates a new httptest.Server OAuth2 provider Access
// Token endpoint. Caller must close the server.
func NewAccessTokenServer(t *testing.T, json string) *httptest.Server {
	return NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		w.Header().Set(contentType, jsonContentType)
		w.Write([]byte(json))
	})
}

// NewTestServeFunc is an adapter to allow the use of ordinary functions as
// httptest.Server's for testing. Caller must close the server.
func NewTestServerFunc(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}
