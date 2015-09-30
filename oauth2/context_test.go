package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestContext_State(t *testing.T) {
	expectedState := "state"
	ctx := WithState(context.Background(), expectedState)
	state, err := StateFromContext(ctx)
	assert.Equal(t, expectedState, state)
	assert.Nil(t, err)
}

func TestContext_MissingState(t *testing.T) {
	state, err := StateFromContext(context.Background())
	assert.Equal(t, "", state)
	if assert.NotNil(t, err) {
		assert.Equal(t, "oauth2: Context missing state value", err.Error())
	}
}

func TestContext_AccessToken(t *testing.T) {
	expectedToken := "access_token"
	ctx := WithAccessToken(context.Background(), expectedToken)
	token, err := AccessTokenFromContext(ctx)
	assert.Equal(t, expectedToken, token)
	assert.Nil(t, err)
}

func TestAccessTokenFromContext_Error(t *testing.T) {
	token, err := AccessTokenFromContext(context.Background())
	assert.Equal(t, "", token)
	if assert.NotNil(t, err) {
		assert.Equal(t, "oauth2: Context missing access token", err.Error())
	}
}
