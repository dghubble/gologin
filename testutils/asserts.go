// Package testutils provides utilities for writing gologin tests.
package testutils

import (
	"io"
	"io/ioutil"
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

// AssertBodyString asserts that a Request Body matches the expected string.
func AssertBodyString(t *testing.T, rc io.ReadCloser, expected string) {
	defer rc.Close()
	if b, err := ioutil.ReadAll(rc); err == nil {
		if string(b) != expected {
			t.Errorf("expected %q, got %q", expected, string(b))
		}
	} else {
		t.Errorf("error reading Body")
	}
}
