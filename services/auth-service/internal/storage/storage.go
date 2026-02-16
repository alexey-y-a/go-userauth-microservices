package storage

import (
	"context"
	"errors"
	"sync"
)

var ErrUserAlreadyExists = errors.New("user already exists")

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
}

type Store interface {
	CreateUser(ctx context.Context, username, email, passwordHash string) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, bool, error)
}

type memoryStore struct {
	mu     sync.RWMutex
	users  map[string]User
	nextID int64
}

func NewMemoryStore() Store {
	return &memoryStore{
		users:  make(map[string]User),
		nextID: 1,
	}
}

func (s *memoryStore) CreateUser(ctx context.Context, username, email, passwordHash string) (User, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exist := s.users[username]
	if exist {
		return User{}, ErrUserAlreadyExists
	}

	id := s.nextID
	s.nextID++

	user := User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	s.users[username] = user

	return user, nil
}

func (s *memoryStore) GetUserByUsername(ctx context.Context, username string) (User, bool, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exist := s.users[username]
	return user, exist, nil
}
