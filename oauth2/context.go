package oauth2

import (
	"fmt"

	"golang.org/x/net/context"
)

// unexported key type prevents collisions
type key int

const (
	accessTokenKey key = iota
	stateKey
)

// WithState returns a copy of ctx that stores the state value.
func WithState(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, stateKey, state)
}

// StateFromContext returns the state value from the ctx.
func StateFromContext(ctx context.Context) (string, error) {
	state, ok := ctx.Value(stateKey).(string)
	if !ok {
		return "", fmt.Errorf("oauth2: Context missing state value")
	}
	return state, nil
}

// WithAccessToken returns a copy of ctx that stores the access token value.
func WithAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, accessTokenKey, accessToken)
}

// AccessTokenFromContext returns the access token value from the ctx.
func AccessTokenFromContext(ctx context.Context) (string, error) {
	accessToken, ok := ctx.Value(accessTokenKey).(string)
	if !ok {
		return "", fmt.Errorf("oauth2: Context missing access token")
	}
	return accessToken, nil
}
