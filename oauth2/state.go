package oauth2

// StateSource returns an OAuth2 State parameter.
type StateSource interface {
	// State returns a OAuth2 state value to prevent CSRF.
	State() string
}

// StateSourceFunc is an adapter to allow a function to be used as a
// StateSource.
type StateSourceFunc func() string

// State returns a OAuth2 state value to prevent CSRF.
func (f StateSourceFunc) State() string {
	return f()
}
