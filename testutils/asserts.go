// Package testutils provides utilities for writing gologin tests.
package testutils

import (
	"net/http"
	"testing"

	"github.com/dghubble/ctxh"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// AssertSuccessNotCalled is a success ContextHandler that fails if called.
func AssertSuccessNotCalled(t *testing.T) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to success ContextHandler")
	}
	return ctxh.ContextHandlerFunc(fn)
}

// AssertFailureNotCalled is a failure ContextHandler that fails if called.
func AssertFailureNotCalled(t *testing.T) ctxh.ContextHandler {
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to failure ContextHandler")
	}
	return ctxh.ContextHandlerFunc(fn)
}
