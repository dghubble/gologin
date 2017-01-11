package digits

import (
	"context"
	"fmt"

	"github.com/dghubble/go-digits/digits"
)

// unexported key type prevents collisions
type key int

const (
	accountKey key = iota
	endpointKey
	headerKey
)

// WithAccount returns a copy of ctx that stores the Digits Account.
func WithAccount(ctx context.Context, account *digits.Account) context.Context {
	return context.WithValue(ctx, accountKey, account)
}

// WithEcho returns a copy of ctx that stores the Digits Echo endpoint and
// header.
func WithEcho(ctx context.Context, endpoint, header string) context.Context {
	ctx = context.WithValue(ctx, endpointKey, endpoint)
	ctx = context.WithValue(ctx, headerKey, header)
	return ctx
}

// AccountFromContext returns the Digits Account from the ctx.
func AccountFromContext(ctx context.Context) (*digits.Account, error) {
	account, ok := ctx.Value(accountKey).(*digits.Account)
	if !ok {
		return nil, fmt.Errorf("digits: Context missing Digits Account")
	}
	return account, nil
}

// EchoFromContext returns the Digits echo endpoint and header from the ctx.
func EchoFromContext(ctx context.Context) (string, string, error) {
	endpoint, okE := ctx.Value(endpointKey).(string)
	header, okH := ctx.Value(headerKey).(string)
	if !okE || !okH {
		return "", "", fmt.Errorf("digits: Context missing Digits echo endpoint or header")
	}
	return endpoint, header, nil
}
