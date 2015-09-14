package gologin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestDefaultErrorHandler(t *testing.T) {
	const expectedMessage = "some error"
	rec := httptest.NewRecorder()
	// should pass through error
	ctx := WithError(context.Background(), fmt.Errorf(expectedMessage))
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)
	DefaultFailureHandler.ServeHTTP(ctx, rec, req)
	assert.Equal(t, expectedMessage+"\n", rec.Body.String())
}
