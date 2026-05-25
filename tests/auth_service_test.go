package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/service"
	"github.com/taskflow/taskflow/internal/utils"
)

func newAuthSvc(t *testing.T) (service.AuthService, *MockUserRepo) {
	t.Helper()
	repo := NewMockUserRepo()
	issuer := auth.NewIssuer("test-secret-xxxxxxxxxxxx", time.Hour, "taskflow-test")
	return service.NewAuthService(repo, issuer), repo
}

func TestAuthService_Register_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.RegisterRequest
		setup   func(r *MockUserRepo)
		wantErr error
	}{
		{
			name: "happy path",
			req:  dto.RegisterRequest{Email: "alice@example.com", Password: "S3cret!Pw", Name: "Alice"},
		},
		{
			name: "duplicate email rejected",
			req:  dto.RegisterRequest{Email: "bob@example.com", Password: "S3cret!Pw", Name: "Bob"},
			setup: func(r *MockUserRepo) {
				_, _ = r.Create(context.Background(), &model.User{
					Email:        "bob@example.com",
					Name:         "Existing Bob",
					PasswordHash: "already-set",
				})
			},
			wantErr: utils.ErrAlreadyExists,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newAuthSvc(t)
			if tc.setup != nil {
				tc.setup(repo)
			}
			res, err := svc.Register(context.Background(), tc.req)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, res.Token)
			require.Equal(t, tc.req.Email, res.User.Email)
			require.True(t, res.ExpiresAt.After(time.Now()))
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	svc, _ := newAuthSvc(t)
	ctx := context.Background()

	_, err := svc.Register(ctx, dto.RegisterRequest{
		Email: "carol@example.com", Password: "S3cret!Pw", Name: "Carol",
	})
	require.NoError(t, err)

	t.Run("valid credentials", func(t *testing.T) {
		res, err := svc.Login(ctx, dto.LoginRequest{Email: "carol@example.com", Password: "S3cret!Pw"})
		require.NoError(t, err)
		require.NotEmpty(t, res.Token)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := svc.Login(ctx, dto.LoginRequest{Email: "carol@example.com", Password: "wrong"})
		require.ErrorIs(t, err, utils.ErrInvalidCreds)
	})

	t.Run("unknown email yields same error", func(t *testing.T) {
		_, err := svc.Login(ctx, dto.LoginRequest{Email: "ghost@example.com", Password: "whatever"})
		require.ErrorIs(t, err, utils.ErrInvalidCreds)
	})
}
