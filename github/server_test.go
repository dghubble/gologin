package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dghubble/gologin/v2/testutils"
)

// newGithubTestServer returns a new httptest.Server which mocks the Github user
// endpoint and a client which proxies requests to the server. The server
// responds with the given json data. The caller must close the server. The
// routePrefix parameter specifies an optional route prefix that should be set
// to the API root route (empty for github.com, "/api/v3" for GHE).
func newGithubTestServer(routePrefix, jsonData string) (*http.Client, *httptest.Server) {
	client, mux, server := testutils.TestServer()
	mux.HandleFunc(routePrefix+"/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, server
}
