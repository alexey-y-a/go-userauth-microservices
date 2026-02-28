package service

import (
	"context"
	"errors"

	"github.com/alexey-y-a/go-userauth-microservices/libs/jwt"
	"github.com/alexey-y-a/go-userauth-microservices/libs/password"
	"github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage"
)

var (
	ErrUserAlreadyExists  = storage.ErrUserAlreadyExists
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	store storage.Store
}

func NewAuthService(store storage.Store) *AuthService {
	return &AuthService{
		store: store,
	}
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type RegisterOutput struct {
	ID       int64
	Username string
	Email    string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {

	_, exists, err := s.store.GetUserByUsername(ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.store.CreateUser(ctx, in.Username, in.Email, hash)
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

type LoginInput struct {
	Username string
	Password string
}

type LoginOutput struct {
	Token string
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	user, exists, err := s.store.GetUserByUsername(ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrInvalidCredentials
	}

	if !password.Compare(user.PasswordHash, in.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := jwt.GenerateAccessToken(user.Username)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{Token: token}, nil
}
