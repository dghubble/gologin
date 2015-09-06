package oauth2

import (
	"fmt"

	"golang.org/x/net/context"
)

// unexported key type prevents collisions
type key int

const (
	accessTokenKey key = 0
)

// WithAccessToken returns a copy of ctx that stores the access token value.
func WithAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, accessTokenKey, accessToken)
}

// AccessTokenFromContext returns the access token value from the ctx.
func AccessTokenFromContext(ctx context.Context) (string, error) {
	accessToken, ok := ctx.Value(accessTokenKey).(string)
	if !ok {
		return "", fmt.Errorf("Context missing access token")
	}
	return accessToken, nil
}
