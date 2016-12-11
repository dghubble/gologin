package oauth1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextRequestToken(t *testing.T) {
	expectedToken := "request_token"
	expectedSecret := "request_secret"
	ctx := WithRequestToken(context.Background(), expectedToken, expectedSecret)
	token, secret, err := RequestTokenFromContext(ctx)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, expectedSecret, secret)
	assert.Nil(t, err)
}

func TestRequestTokenFromContext_Error(t *testing.T) {
	token, secret, err := RequestTokenFromContext(context.Background())
	assert.Equal(t, "", token)
	assert.Equal(t, "", secret)
	if assert.NotNil(t, err) {
		assert.Equal(t, "oauth1: Context missing request token or secret", err.Error())
	}
}

func TestContextAccessToken(t *testing.T) {
	expectedToken := "access_token"
	expectedSecret := "access_secret"
	ctx := WithAccessToken(context.Background(), expectedToken, expectedSecret)
	token, secret, err := AccessTokenFromContext(ctx)
	assert.Equal(t, expectedToken, token)
	assert.Equal(t, expectedSecret, secret)
	assert.Nil(t, err)
}

func TestAccessTokenFromContext_Error(t *testing.T) {
	token, secret, err := AccessTokenFromContext(context.Background())
	assert.Equal(t, "", token)
	assert.Equal(t, "", secret)
	if assert.NotNil(t, err) {
		assert.Equal(t, "oauth1: Context missing access token or secret", err.Error())
	}
}
