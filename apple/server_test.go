package apple

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dghubble/gologin/v2/testutils"
)

// newAppleTestServer returns a new httptest.Server which mocks the AppleID
// server endpoint and a client which proxies requests to the server. The server
// responds with the given json data. The caller must close the server.
func newAppleTestServer(keysJSONData string) (*http.Client, *httptest.Server) {
	client, mux, server := testutils.TestServer()

	mux.HandleFunc(AppleBaseURL+"/auth/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, keysJSONData)
	})

	return client, server
}
