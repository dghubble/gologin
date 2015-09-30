package google

import (
	"fmt"

	google "github.com/google/google-api-go-client/oauth2/v2"
	"golang.org/x/net/context"
)

// unexported key type prevents collisions
type key int

const (
	userKey key = iota
)

// WithUser returns a copy of ctx that stores the Google Userinfoplus.
func WithUser(ctx context.Context, user *google.Userinfoplus) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Google Userinfoplus from the ctx.
func UserFromContext(ctx context.Context) (*google.Userinfoplus, error) {
	user, ok := ctx.Value(userKey).(*google.Userinfoplus)
	if !ok {
		return nil, fmt.Errorf("google: Context missing Google User")
	}
	return user, nil
}
