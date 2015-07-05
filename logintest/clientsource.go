package logintest

import (
	"net/http"
)

// StubClientSource is a source for a fixed http.Client.
type StubClientSource struct {
	client *http.Client
}

// NewStubClientSource returns a stubClientSource which always returns the
// given http client.
func NewStubClientSource(client *http.Client) *StubClientSource {
	return &StubClientSource{
		client: client,
	}
}

// GetClient returns the fixed http.Client.
func (s *StubClientSource) GetClient(token, tokenSecret string) *http.Client {
	return s.client
}
