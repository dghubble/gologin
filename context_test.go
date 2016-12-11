package gologin

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextError(t *testing.T) {
	expectedError := fmt.Errorf("some error")
	ctx := WithError(context.Background(), expectedError)
	err := ErrorFromContext(ctx)
	assert.Equal(t, expectedError, err)
}

func TestErrorFromContext_Error(t *testing.T) {
	err := ErrorFromContext(context.Background())
	if assert.NotNil(t, err) {
		assert.Equal(t, "Context missing error value", err.Error())
	}
}
