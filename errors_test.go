package gologin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultFailureHandler(t *testing.T) {
	expectedError := fmt.Errorf("some error")
	ctx := WithError(context.Background(), expectedError)
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)
	w := httptest.NewRecorder()
	DefaultFailureHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// assert that error message was passed through
	assert.Equal(t, expectedError.Error()+"\n", w.Body.String())
}
