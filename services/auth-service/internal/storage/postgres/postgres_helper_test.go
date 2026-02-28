//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const postgresImage = "postgres:16-alpine"

type PostgresContainer struct {
	Container testcontainers.Container
	DSN       string
}

func startPostgres(t *testing.T) *PostgresContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	req := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "auth_user",
			"POSTGRES_PASSWORD": "auth_pass",
			"POSTGRES_DB":       "auth_db",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start postgres")

	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	require.NoError(t, err, "failed to get container host")

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err, "failed to get mapped port")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		"auth_user",
		"auth_pass",
		host,
		mappedPort.Port(),
		"auth_db",
	)

	return &PostgresContainer{
		Container: container,
		DSN:       dsn,
	}
}
