package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateJWT(t *testing.T) {
	signingKey := []byte("test-signing-key")
	username := "testuser"

	t.Run("Valid token creation", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)
		token, err := CreateJWT(username, signingKey, expiresAt)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Expired token creation", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
		token, err := CreateJWT(username, signingKey, expiresAt)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ParseJWT(token, signingKey)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestParseJWT(t *testing.T) {
	signingKey := []byte("test-signing-key")
	username := "testuser"
	expiresAt := time.Now().Add(1 * time.Hour)

	validToken, _ := CreateJWT(username, signingKey, expiresAt)

	t.Run("Valid token parsing", func(t *testing.T) {
		claims, err := ParseJWT(validToken, signingKey)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, username, claims["username"])
	})

	t.Run("Invalid signing key", func(t *testing.T) {
		wrongKey := []byte("wrong-signing-key")
		claims, err := ParseJWT(validToken, wrongKey)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Invalid token format", func(t *testing.T) {
		claims, err := ParseJWT("invalid-token", signingKey)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Expired token", func(t *testing.T) {
		expiredToken, _ := CreateJWT(username, signingKey, time.Now().Add(-1*time.Hour))
		claims, err := ParseJWT(expiredToken, signingKey)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
