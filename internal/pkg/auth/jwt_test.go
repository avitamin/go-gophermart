package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_Generate(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)

	token, err := manager.Generate(123)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_Verify(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	userID := int64(123)

	token, err := manager.Generate(userID)
	require.NoError(t, err)

	claims, err := manager.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestJWTManager_VerifyExpired(t *testing.T) {
	manager := NewJWTManager("test-secret", -time.Hour)

	token, err := manager.Generate(123)
	require.NoError(t, err)

	_, err = manager.Verify(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTManager_VerifyInvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)

	_, err := manager.Verify("invalid-token")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTManager_VerifyWrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-1", time.Hour)
	manager2 := NewJWTManager("secret-2", time.Hour)

	token, err := manager1.Generate(123)
	require.NoError(t, err)

	_, err = manager2.Verify(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}
