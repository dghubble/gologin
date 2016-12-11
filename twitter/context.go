package twitter

import (
	"context"
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
)

// unexported key type prevents collisions
type key int

const (
	userKey key = iota
)

// WithUser returns a copy of ctx that stores the Twitter User.
func WithUser(ctx context.Context, user *twitter.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Twitter User from the ctx.
func UserFromContext(ctx context.Context) (*twitter.User, error) {
	user, ok := ctx.Value(userKey).(*twitter.User)
	if !ok {
		return nil, fmt.Errorf("twitter: Context missing Twitter User")
	}
	return user, nil
}
