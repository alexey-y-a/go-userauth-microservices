package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage"
)

func (s *Store) CreateUser(username, email, passwordHash string) (storage.User, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    const query = `
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, username, email, password_hash;`

    var user storage.User

    err := s.db.QueryRowContext(ctx, query, username, email, passwordHash).
        Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
    if err != nil {
        return storage.User{}, fmt.Errorf("insert user: %w", err)
    }

    return user, nil
}

func (s *Store) GetUserByUsername(username string) (storage.User, bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    const query = `
SELECT id, username, email, password_hash
FROM users
WHERE username = $1;`

    var user storage.User

    err := s.db.QueryRowContext(ctx, query, username).
        Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
    if err != nil {
        if err == sql.ErrNoRows {
            return storage.User{}, false, nil
        }

        return storage.User{}, false, fmt.Errorf("select user by username: %w", err)
    }

    return user, true, nil
}
