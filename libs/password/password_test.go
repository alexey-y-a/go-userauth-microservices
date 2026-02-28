package password

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndCompare(t *testing.T) {
	password := "secret123"

	hash, err := Hash(password)
	require.NoError(t, err, "hash must not return error")
	require.NotEmpty(t, hash, "hash must not be empty")
	require.NotEqual(t, password, hash, "hash must differ from plaintext")

	require.True(t, Compare(hash, password), "compare must be true for valid password")
	require.False(t, Compare(hash, "wrong-password"), "compare must be false for invalid password")
}
