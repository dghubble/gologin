package oauth1

import (
	"context"
	"fmt"
)

// unexported key type prevents collisions
type key int

const (
	requestTokenKey key = iota
	requestSecretKey
	accessTokenKey
	accessSecretKey
)

// WithRequestToken returns a copy of ctx that stores the request token and
// secret values.
func WithRequestToken(ctx context.Context, requestToken, requestSecret string) context.Context {
	ctx = context.WithValue(ctx, requestTokenKey, requestToken)
	ctx = context.WithValue(ctx, requestSecretKey, requestSecret)
	return ctx
}

// RequestTokenFromContext returns the request token and secret from the ctx.
func RequestTokenFromContext(ctx context.Context) (string, string, error) {
	requestToken, okT := ctx.Value(requestTokenKey).(string)
	requestSecret, okS := ctx.Value(requestSecretKey).(string)
	if !okT || !okS {
		return "", "", fmt.Errorf("oauth1: Context missing request token or secret")
	}
	return requestToken, requestSecret, nil
}

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
		return "", "", fmt.Errorf("oauth1: Context missing access token or secret")
	}
	return accessToken, accessSecret, nil
}
