package testutils

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestServer returns an http Client, ServeMux, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func TestServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	transport := &RewriteTransport{&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}}
	client := &http.Client{Transport: transport}
	return client, mux, server
}

// UnauthorizedTestServer returns a http.Server which always returns a 401
// Unauthorized response and a client which proxies to it. The caller must
// close the test server.
func UnauthorizedTestServer() (*http.Client, *httptest.Server) {
	client, mux, server := TestServer()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	return client, server
}

// RewriteTransport rewrites https requests to http to avoid TLS cert issues
// during testing.
type RewriteTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

// NewErrorServer returns a new httptest.Server endpoint which responds with
// the given error message and code. Caller must close the server.
func NewErrorServer(t *testing.T, message string, code int) *httptest.Server {
	return NewTestServerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		http.Error(w, message, code)
	})
}

// NewTestServerFunc is an adapter to allow the use of ordinary functions as
// httptest.Server's for testing. Caller must close the server.
func NewTestServerFunc(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}
