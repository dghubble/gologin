package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v64/github"
)

// unexported key type prevents collisions
type key int

const (
	userKey key = iota
)

// WithUser returns a copy of ctx that stores the GitHub User.
func WithUser(ctx context.Context, user *github.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the GitHub User from the ctx.
func UserFromContext(ctx context.Context) (*github.User, error) {
	user, ok := ctx.Value(userKey).(*github.User)
	if !ok {
		return nil, fmt.Errorf("github: Context missing GitHub User")
	}
	return user, nil
}
