package logintest

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// ErrorOnFailure is an ErrorHandler which asserts that it is not called.
func ErrorOnFailure(t *testing.T) func(w http.ResponseWriter, err error, code int) {
	failure := func(w http.ResponseWriter, err error, code int) {
		t.Errorf("unexpected call to failure, %v %d", err, code)
	}
	return failure
}

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

// AssertBodyString asserts that a Request Body matches the expected string.
func AssertBodyString(t *testing.T, rc io.ReadCloser, expected string) {
	defer rc.Close()
	if b, err := ioutil.ReadAll(rc); err == nil {
		if string(b) != expected {
			t.Errorf("expected %q, got %q", expected, string(b))
		}
	} else {
		t.Errorf("error reading Body")
	}
}
