//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage"
	pg "github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage/postgres"
)

func TestPostgresStore_CreateAndGetUser(t *testing.T) {
	pgContainer := startPostgres(t)

	cfg := pg.Config{DSN: pgContainer.DSN}
	store, err := pg.NewStore(cfg)
	require.NoError(t, err, "NewStore must succeed")
	t.Cleanup(func() {
		_ = store.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user1, err := store.CreateUser(ctx, "alice", "alice@example.com", "hash-1")
	require.NoError(t, err, "CreateUser must succeed for new username")
	require.NotZero(t, user1.ID)
	require.Equal(t, "alice", user1.Username)
	require.Equal(t, "alice@example.com", user1.Email)
	require.Equal(t, "hash-1", user1.PasswordHash)

	_, err = store.CreateUser(ctx, "alice", "alice2@example.com", "hash-2")
	require.Error(t, err, "CreateUser must fail for duplicate username")
	require.ErrorIs(t, err, storage.ErrUserAlreadyExists)

	got, found, err := store.GetUserByUsername(ctx, "alice")
	require.NoError(t, err, "GetUserByUsername must not fail")
	require.True(t, found, "user must be found")
	require.Equal(t, user1.ID, got.ID)
	require.Equal(t, user1.Username, got.Username)
	require.Equal(t, user1.Email, got.Email)
	require.Equal(t, user1.PasswordHash, got.PasswordHash)

	_, found, err = store.GetUserByUsername(ctx, "bob")
	require.NoError(t, err)
	require.False(t, found, "user must not be found")
}
