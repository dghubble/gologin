package gologin

import (
	"fmt"

	"golang.org/x/net/context"
)

// unexported key type prevents collisions
type key int

const (
	errorKey key = 0
)

// WithError returns a copy of ctx that stores the given error value.
func WithError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

// ErrorFromContext returns the error value from the ctx or an error that the
// context was missing an error value.
func ErrorFromContext(ctx context.Context) error {
	err, ok := ctx.Value(errorKey).(error)
	if !ok {
		return fmt.Errorf("Context missing error value")
	}
	return err
}
