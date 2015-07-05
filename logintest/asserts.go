package logintest

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

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

// ErrorOnFailure is an ErrorHandler which asserts that it is not called.
func ErrorOnFailure(t *testing.T) func(w http.ResponseWriter, err error, code int) {
	failure := func(w http.ResponseWriter, err error, code int) {
		t.Errorf("unexpected call to failure, %v %d", err, code)
	}
	return failure
}
