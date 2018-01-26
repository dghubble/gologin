package github

import (
	"context"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestContextUser(t *testing.T) {
	expectedUser := &github.User{
		ID:   github.Int64(917408),
		Name: github.String("Github User"),
	}
	ctx := WithUser(context.Background(), expectedUser)
	user, err := UserFromContext(ctx)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
}

func TestContextUser_Error(t *testing.T) {
	user, err := UserFromContext(context.Background())
	assert.Nil(t, user)
	if assert.NotNil(t, err) {
		assert.Equal(t, "github: Context missing Github User", err.Error())
	}
}
