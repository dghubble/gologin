package bitbucket

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/jbcjorge/gologin/testutils"
)

// newBitbucketTestServer returns a new httptest.Server which mocks the
// Bitbucket user endpoint and a client which proxies requests to the server.
// The server responds with the given json data. The caller must close the
// server.
func newBitbucketTestServer(jsonData string) (*http.Client, *httptest.Server) {
	client, mux, server := testutils.TestServer()
	mux.HandleFunc("/api/2.0/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, server
}
