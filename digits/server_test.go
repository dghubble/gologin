package digits

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/jbcjorge/gologin/testutils"
)

// newDigitsTestServer returns a new httptest.Server which mocks the Digits
// accounts endpoint and a client which proxies requests to the server. The
// server responds with the given json data. The caller must close the server.
func newDigitsTestServer(jsonData string) (*http.Client, *http.ServeMux, *httptest.Server) {
	client, mux, server := testutils.TestServer()
	mux.HandleFunc("/1.1/sdk/account.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, jsonData)
	})
	return client, mux, server
}
