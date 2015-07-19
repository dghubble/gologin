package oauth2

// StateSource returns an OAuth2 State parameter.
type StateSource interface {
	State() string
}

type StateSourceFunc func() string

func (f StateSourceFunc) State() string {
	return f()
}
