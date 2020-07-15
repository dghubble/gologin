package google

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dghubble/gologin/v2/testutils"
)

// newGoogleTestServer returns a new httptest.Server which mocks the Google
// Userinfo endpoint and a client which proxies requests to the server.
// The server responds with the given json data. The caller must close the
// server.
func newGoogleTestServer(jsonData string) (*http.Client, *httptest.Server) {
	client, mux, server := testutils.TestServer()
	mux.HandleFunc("/oauth2/v2/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, server
}
