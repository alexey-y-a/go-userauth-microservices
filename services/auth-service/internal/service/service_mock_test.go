package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/alexey-y-a/go-userauth-microservices/services/auth-service/internal/storage"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) CreateUser(ctx context.Context, username, email, passwordHash string) (storage.User, error) {
	args := m.Called(ctx, username, email, passwordHash)

	user, _ := args.Get(0).(storage.User)
	return user, args.Error(1)

}
func (m *MockStore) GetUserByUsername(ctx context.Context, username string) (storage.User, bool, error) {
	args := m.Called(ctx, username)

	user, _ := args.Get(0).(storage.User)
	found, _ := args.Get(1).(bool)
	return user, found, args.Error(2)
}

func TestAuthService_Register_WithMockStore(t *testing.T) {
	ctx := context.Background()

	type args struct {
		in RegisterInput
	}

	dbErr := errors.New("db error")

	tests := []struct {
		name       string
		setupMock  func(m *MockStore)
		args       args
		wantErr    error
		wantUserID int64
	}{
		{
			name: "ok - new user",
			setupMock: func(m *MockStore) {
				m.On("GetUserByUsername", mock.Anything, "alice").
					Return(storage.User{}, false, nil).
					Once()

				m.On("CreateUser", mock.Anything, "alice", "alice@example.com", mock.AnythingOfType("string")).
					Return(storage.User{
						ID:           1,
						Username:     "alice",
						Email:        "alice@example.com",
						PasswordHash: "some-hash",
					}, nil).
					Once()
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice@example.com",
					Password: "secret123",
				},
			},
			wantErr:    nil,
			wantUserID: 1,
		},
		{
			name: "user already exists",
			setupMock: func(m *MockStore) {
				m.On("GetUserByUsername", mock.Anything, "alice").
					Return(storage.User{ID: 1, Username: "alice"}, true, nil).
					Once()
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice@example.com",
					Password: "secret123",
				},
			},
			wantErr:    ErrUserAlreadyExists,
			wantUserID: 0,
		},
		{
			name: "GetUserByUsername returns error",
			setupMock: func(m *MockStore) {
				m.On("GetUserByUsername", mock.Anything, "alice").
					Return(storage.User{}, false, dbErr).
					Once()
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice@example.com",
					Password: "secret123",
				},
			},
			wantErr:    dbErr,
			wantUserID: 0,
		},
		{
			name: "CreateUser returns error",
			setupMock: func(m *MockStore) {
				m.On("GetUserByUsername", mock.Anything, "alice").
					Return(storage.User{}, false, nil).
					Once()

				m.On("CreateUser", mock.Anything, "alice", "alice@example.com", mock.AnythingOfType("string")).
					Return(storage.User{}, dbErr).
					Once()
			},
			args: args{
				in: RegisterInput{
					Username: "alice",
					Email:    "alice@example.com",
					Password: "secret123",
				},
			},
			wantErr:    dbErr,
			wantUserID: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockStore{}
			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			svc := NewAuthService(mockStore)

			out, err := svc.Register(ctx, tt.args.in)

			if tt.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, out)
				require.Equal(t, tt.wantUserID, out.ID)
				require.Equal(t, tt.args.in.Username, out.Username)
				require.Equal(t, tt.args.in.Email, out.Email)
			} else {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrUserAlreadyExists) {
					require.ErrorIs(t, err, ErrUserAlreadyExists)
				} else {
					require.EqualError(t, err, tt.wantErr.Error())
				}
				require.Nil(t, out)
			}

			mockStore.AssertExpectations(t)
		})
	}
}
