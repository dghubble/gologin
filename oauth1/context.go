package oauth1

import (
	"fmt"

	"golang.org/x/net/context"
)

// unexported key type prevents collisions
type key int

const (
	accessTokenKey key = iota
	accessSecretKey
)

// WithAccessToken returns a copy of ctx that stores the access token and
// secret values.
func WithAccessToken(ctx context.Context, accessToken, accessSecret string) context.Context {
	ctx = context.WithValue(ctx, accessTokenKey, accessToken)
	ctx = context.WithValue(ctx, accessSecretKey, accessSecret)
	return ctx
}

// AccessTokenFromContext returns the access token and secret from the ctx.
func AccessTokenFromContext(ctx context.Context) (string, string, error) {
	accessToken, okT := ctx.Value(accessTokenKey).(string)
	accessSecret, okS := ctx.Value(accessSecretKey).(string)
	if !okT || !okS {
		return "", "", fmt.Errorf("Context missing access token or secret")
	}
	return accessToken, accessSecret, nil
}
