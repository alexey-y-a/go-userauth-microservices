package requestid

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithContextAndFromContext(t *testing.T) {
	ctx := context.Background()
	const id = "test-request-id"

	ctxWithID := WithContext(ctx, id)

	got, ok := FromContext(ctxWithID)
	require.True(t, ok, "FromContext must report that id exists")
	require.Equal(t, id, got, "FromContext must return the same id we stored")

	_, ok = FromContext(context.Background())
	require.False(t, ok, "FromContext must return false when id is absent")
}

func TestGetOrGenerate_UsesHeader(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	const id = "header-request-id"
	req.Header.Set(HeaderName, id)

	got := GetOrGenerate(req)
	require.Equal(t, id, got, "GetOrGenerate must return id from header when present")
}

func TestGetOrGenerate_GeneratesNewID(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	got := GetOrGenerate(req)
	require.NotEmpty(t, got, "GetOrGenerate must generate non-empty id when header is missing")

	req2, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	got2 := GetOrGenerate(req2)
	require.NotEmpty(t, got2)
	require.NotEqual(t, got, got2, "two generated ids should deffer")
}
