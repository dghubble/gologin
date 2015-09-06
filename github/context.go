package github

import (
	"fmt"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
)

// unexported key type prevents collisions with keys from other packages
type key int

const (
	userKey  key = 0
	errorKey key = 1
)

// WithAccessToken returns a copy of ctx that stores the access token value.
func WithUser(ctx context.Context, user *github.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Github User from the ctx.
func UserFromContext(ctx context.Context) (*github.User, error) {
	user, ok := ctx.Value(userKey).(*github.User)
	if !ok {
		return nil, fmt.Errorf("Context missing Github user")
	}
	return user, nil
}
