package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alexey-y-a/go-userauth-microservices/libs/jwt"
	"github.com/alexey-y-a/go-userauth-microservices/libs/password"
	"github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage"
)

type fakeStore struct {
	users       map[string]storage.User
	errOnGet    error
	errOnCreate error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users: make(map[string]storage.User),
	}
}

func (s *fakeStore) CreateUser(ctx context.Context, username, email, passwordHash string) (storage.User, error) {
	if s.errOnCreate != nil {
		return storage.User{}, s.errOnCreate
	}
	id := int64(len(s.users) + 1)

	user := storage.User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}
	s.users[username] = user
	return user, nil
}

func (s *fakeStore) GetUserByUsername(ctx context.Context, username string) (storage.User, bool, error) {
	if s.errOnGet != nil {
		return storage.User{}, false, s.errOnGet
	}
	u, ok := s.users[username]
	return u, ok, nil
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		store *fakeStore
	}

	type args struct {
		in RegisterInput
	}

	dbErr := errors.New("db error")

	tests := []struct {
		name      string
		setup     func(f *fields)
		args      args
		wantErr   error
		wantEmail string
	}{
		{
			name: "ok - new user",
			setup: func(f *fields) {
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice@example.com",
					Password: "secret123",
				},
			},
			wantErr:   nil,
			wantEmail: "alice@example.com",
		},
		{
			name: "user already exists",
			setup: func(f *fields) {
				f.store.users["alice"] = storage.User{
					ID:           1,
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: "hash",
				}
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice2@example.com",
					Password: "secret123",
				},
			},
			wantErr:   ErrUserAlreadyExists,
			wantEmail: "",
		},
		{
			name: "GetUserByUsername returns error",
			setup: func(f *fields) {
				f.store.errOnGet = dbErr
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice@example.com",
					Password: "secret123",
				},
			},
			wantErr:   dbErr,
			wantEmail: "",
		},
		{
			name: "CreateUser returns error",
			setup: func(f *fields) {
				f.store.errOnCreate = dbErr
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice@example.com",
					Password: "secret123",
				},
			},
			wantErr:   dbErr,
			wantEmail: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := &fields{
				store: newFakeStore(),
			}
			if tt.setup != nil {
				tt.setup(f)
			}

			svc := NewAuthService(f.store)

			out, err := svc.Register(ctx, tt.args.in)

			if tt.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, out)
				require.Equal(t, tt.wantEmail, out.Email)
				require.NotZero(t, out.ID)

				stored, ok := f.store.users[tt.args.in.Username]
				require.True(t, ok)
				require.NotEmpty(t, stored.PasswordHash)
				require.NotEqual(t, tt.args.in.Password, stored.PasswordHash)
				require.True(t, password.Compare(stored.PasswordHash, tt.args.in.Password))
			} else {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrUserAlreadyExists) || errors.Is(tt.wantErr, storage.ErrUserAlreadyExists) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
				require.Nil(t, out)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		store *fakeStore
	}

	type args struct {
		in LoginInput
	}

	dbErr := errors.New("db error")

	hashed, err := password.Hash("secret123")
	require.NoError(t, err)

	t.Setenv("JWT_SECRET", "test-secret")

	tests := []struct {
		name    string
		setup   func(f *fields)
		args    args
		wantErr error
	}{
		{
			name: "ok - valid credentials",
			setup: func(f *fields) {
				f.store.users["alice"] = storage.User{
					ID:           1,
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: hashed,
				}
			},
			args: args{
				in: LoginInput{
					Username: "alice",
					Password: "secret123",
				},
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			setup: func(f *fields) {
			},
			args: args{
				in: LoginInput{
					Username: "bob",
					Password: "secret123",
				},
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			setup: func(f *fields) {
				f.store.users["alice"] = storage.User{
					ID:           1,
					Username:     "alice",
					Email:        "alice@example.com",
					PasswordHash: hashed,
				}
			},
			args: args{
				in: LoginInput{
					Username: "alice",
					Password: "wrong-password",
				},
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "GetUserByUsername returns error",
			setup: func(f *fields) {
				f.store.errOnGet = dbErr
			},
			args: args{
				in: LoginInput{
					Username: "alice",
					Password: "secret123",
				},
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := &fields{
				store: newFakeStore(),
			}
			if tt.setup != nil {
				tt.setup(f)
			}

			svc := NewAuthService(f.store)

			out, err := svc.Login(ctx, tt.args.in)

			if tt.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, out)
				require.NotEmpty(t, out.Token)

				claims, err := jwt.ParseToken(out.Token)
				require.NoError(t, err)
				require.Equal(t, "alice", claims.UserID)
			} else {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrInvalidCredentials) {
					require.ErrorIs(t, err, ErrInvalidCredentials)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
				require.Nil(t, out)
			}
		})
	}
}
