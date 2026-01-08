package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptHasher_Hash(t *testing.T) {
	hasher := NewBcryptHasher()

	hash, err := hasher.Hash("password123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "password123", hash)
}

func TestBcryptHasher_Compare(t *testing.T) {
	hasher := NewBcryptHasher()

	password := "securePassword123!"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	assert.True(t, hasher.Compare(password, hash))
	assert.False(t, hasher.Compare("wrongPassword", hash))
}

func TestBcryptHasher_DifferentHashesForSamePassword(t *testing.T) {
	hasher := NewBcryptHasher()

	password := "samePassword"
	hash1, err := hasher.Hash(password)
	require.NoError(t, err)

	hash2, err := hasher.Hash(password)
	require.NoError(t, err)

	// Bcrypt generates different hashes for the same password (due to salt)
	assert.NotEqual(t, hash1, hash2)

	// But both should validate correctly
	assert.True(t, hasher.Compare(password, hash1))
	assert.True(t, hasher.Compare(password, hash2))
}
