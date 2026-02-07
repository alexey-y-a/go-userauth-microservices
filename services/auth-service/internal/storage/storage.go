package storage

import (
	"errors"
	"sync"
)

var ErrUserAlreadyExists = errors.New("user already exists")

type User struct {
    ID int64
    Username string
    Email string
    PasswordHash string
}

type Store interface {
    CreateUser(username, email, passwordHash string) (User, error)
    GetUserByUsername(username string) (User, bool, error)
}

type memoryStore struct {
    mu sync.RWMutex
    users map[string]User
    nextID int64
}

func NewMemoryStore() Store {
	return &memoryStore{
		users:  make(map[string]User),
		nextID: 1,
	}
}

func (s *memoryStore) CreateUser(username, email, passwordHash string) (User, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    _, exist := s.users[username]
    if exist {
        return User{}, ErrUserAlreadyExists
    }

    id := s.nextID
    s.nextID++

    user := User {
        ID: id,
        Username: username,
        Email: email,
        PasswordHash: passwordHash,
    }

    s.users[username] = user

    return user, nil
}

func (s *memoryStore) GetUserByUsername(username string) (User, bool, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    user, exist := s.users[username]
    return user, exist, nil
}