package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("testpassword")
	require.NoError(t, err)
	require.NotNil(t, hash)
	err = CheckPasswordHash(hash, "testpassword")
	require.NoError(t, err)
}
