package storage

import "context"

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
}

type UserStore interface {
	GetUserByID(ctx context.Context, id int64) (User, bool, error)
	GetUserByUsername(ctx context.Context, username string) (User, bool, error)
}
