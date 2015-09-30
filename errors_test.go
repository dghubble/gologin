package gologin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestDefaultFailureHandler(t *testing.T) {
	expectedError := fmt.Errorf("some error")
	ctx := WithError(context.Background(), expectedError)
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)
	w := httptest.NewRecorder()
	DefaultFailureHandler.ServeHTTP(ctx, w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// assert that error message was passed through
	assert.Equal(t, expectedError.Error()+"\n", w.Body.String())
}
