package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alexey-y-a/go-userauth-microservices/services/user-service/internal/storage"
)

func (s *Store) GetUserByID(ctx context.Context, id int64) (storage.User, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	const query = `
SELECT id, username, email, password_hash
FROM users
WHERE id = $1
`

	var user storage.User

	err := s.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return storage.User{}, false, nil
		}
		return storage.User{}, false, fmt.Errorf("select user by id: %w", err)
	}

	return user, true, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (storage.User, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	const query = `
SELECT id, username, email, password_hash
FROM users
WHERE username = $1
`
	var user storage.User

	err := s.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return storage.User{}, false, nil
		}
		return storage.User{}, false, fmt.Errorf("select user by username: %w", err)
	}

	return user, true, nil
}
