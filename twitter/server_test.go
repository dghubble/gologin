package twitter

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dghubble/gologin/v2/testutils"
)

// newTwitterVerifyServer returns a new httptest.Server which mocks the Twitter
// verify credentials endpoint and a client which proxies requests to the
// server. The server responds with the given json data. The caller must close
// the server.
func newTwitterVerifyServer(jsonData string) (*http.Client, *http.ServeMux, *httptest.Server) {
	client, mux, server := testutils.TestServer()
	mux.HandleFunc("/1.1/account/verify_credentials.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, mux, server
}
