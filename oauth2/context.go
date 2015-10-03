package oauth2

import (
	"fmt"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
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

// WithAccessToken returns a copy of ctx that stores the access Token.
func WithAccessToken(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, accessTokenKey, token)
}

// AccessTokenFromContext returns the access Token from the ctx.
func AccessTokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	token, ok := ctx.Value(accessTokenKey).(*oauth2.Token)
	if !ok {
		return nil, fmt.Errorf("oauth2: Context missing access Token")
	}
	return token, nil
}
