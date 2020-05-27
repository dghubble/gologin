package google

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	google "google.golang.org/api/oauth2/v2"
)

func TestContextUser(t *testing.T) {
	expectedUserinfoplus := &google.Userinfo{Id: "42", Name: "Google User"}
	ctx := WithUser(context.Background(), expectedUserinfoplus)
	user, err := UserFromContext(ctx)
	assert.Equal(t, expectedUserinfoplus, user)
	assert.Nil(t, err)
}

func TestContextUser_Error(t *testing.T) {
	user, err := UserFromContext(context.Background())
	assert.Nil(t, user)
	if assert.NotNil(t, err) {
		assert.Equal(t, "google: Context missing Google User", err.Error())
	}
}
