package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/taskflow/taskflow/internal/dto"
	"github.com/taskflow/taskflow/internal/model"
	"github.com/taskflow/taskflow/internal/service"
	"github.com/taskflow/taskflow/internal/utils"
	"github.com/taskflow/taskflow/internal/worker"
)

func noopPool(t *testing.T) *worker.Pool {
	t.Helper()
	p := worker.NewPool(1, 16, zap.NewNop(), func(ctx context.Context, e worker.Event) error {
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)
	t.Cleanup(func() {
		cancel()
		p.Stop()
	})
	return p
}

func TestProjectService_CreateAndAccess(t *testing.T) {
	repo := NewMockProjectRepo()
	svc := service.NewProjectService(repo, noopPool(t))
	ctx := context.Background()

	p, err := svc.Create(ctx, 1, dto.CreateProjectRequest{Name: "Apollo", Description: "moon"})
	require.NoError(t, err)
	require.NotZero(t, p.ID)
	require.Equal(t, "Apollo", p.Name)

	got, err := svc.Get(ctx, 1, p.ID)
	require.NoError(t, err)
	require.Equal(t, p.ID, got.ID)

	_, err = svc.Get(ctx, 99, p.ID)
	require.ErrorIs(t, err, utils.ErrForbidden)
}

func TestProjectService_AddMember_RoleGate(t *testing.T) {
	tests := []struct {
		name      string
		actorRole model.ProjectRole
		wantErr   error
	}{
		{"owner can add", model.RoleOwner, nil},
		{"admin can add", model.RoleAdmin, nil},
		{"member cannot add", model.RoleMember, utils.ErrForbidden},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMockProjectRepo()
			svc := service.NewProjectService(repo, noopPool(t))
			ctx := context.Background()

			_, _ = svc.Create(ctx, 1, dto.CreateProjectRequest{Name: "Test"})
			require.NoError(t, repo.AddMember(ctx, 1, 2, tc.actorRole))

			err := svc.AddMember(ctx, 2, 1, dto.AddMemberRequest{UserID: 3, Role: "member"})
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestProjectService_DeleteRequiresOwner(t *testing.T) {
	repo := NewMockProjectRepo()
	svc := service.NewProjectService(repo, noopPool(t))
	ctx := context.Background()

	p, err := svc.Create(ctx, 1, dto.CreateProjectRequest{Name: "Owned"})
	require.NoError(t, err)

	require.NoError(t, repo.AddMember(ctx, p.ID, 2, model.RoleAdmin))
	require.ErrorIs(t, svc.Delete(ctx, 2, p.ID), utils.ErrForbidden)

	require.NoError(t, svc.Delete(ctx, 1, p.ID))
}
